mod centrifugo;
mod rpc_tonic_impl;

use rpc_tonic_impl::CentrifugoClient;

use crossbeam_channel::{Sender, unbounded};
use tokio::task::JoinHandle;
use tokio::runtime::{Builder, Runtime};

use std::mem;

type Result<T> = std::result::Result<T, Box<dyn std::error::Error>>;

pub(crate) struct Publication {
    channel: String,
    data: Vec<u8>,
}

pub(crate) trait Client {
    async fn publish(&mut self, batch: Vec<Publication>) -> anyhow::Result<()>;
}

pub(crate) struct Options {
    address: String,
}

pub(crate) struct App {
    runtime: Runtime,
    sender: Option<Sender<Publication>>,
    worker: Option<JoinHandle<()>>,
}

impl App {
    pub(crate) fn new(options: Options) -> Result<Self> {
        let rt = Builder::new_multi_thread()
            .enable_io()
            .enable_time()
            .worker_threads(1)
            .thread_name("pubsub-rt")
            .build()?;

        let (s, r) = unbounded::<Publication>();

        let mut client = rt.handle().block_on(async {
            // <new> should be called inside runtime (tonic impl require it)
            return CentrifugoClient::new(options.address.as_str()).await
        })?;

        let handle = rt.handle().spawn(async move {
            loop {
                let count = r.len();
                let mut batch = Vec::new();
                loop {
                    match r.recv() {
                        Ok(msg) => {
                            batch.push(msg);
                            if batch.len() >= count {
                                break
                            }
                        },
                        Err(_) => {
                            break
                        },
                    }
                }

                if let Err(e) = client.publish(batch).await {
                    log::error!("publication failed with error: {}", e);
                }
            }
        });

        Ok(Self{
            runtime: rt,
            sender: Some(s),
            worker: Some(handle),
        })
    }

    pub fn send(&self, msg: Publication) -> Result<()> {
        let s = self.sender.as_ref().unwrap();
        match s.send(msg) {
            Ok(v) => Ok(v),
            Err(e) => Err(Box::new(e)),
        }
    }
}

impl Drop for App {
    fn drop(&mut self) {
        if let Some(s) = self.sender.take() {
            // drop sender so receiver loop can be also exited
            mem::drop(s);
        }

        if let Some(h) = self.worker.take() {
            let _ = self.runtime.handle().block_on(h);
        }
    }
}

mod ffi {
    use super::{App, Options, Publication};
    use tarantool::{error::TarantoolErrorCode, ffi};
    use std::ffi::{CStr, c_char, c_int};
    use std::ptr;

    #[repr(C)]
    struct COptions {
        address: *const c_char,
    }

    #[no_mangle]
    unsafe extern "C" fn ffi_pubsub_start(copts: *const COptions) -> *mut App {
        let addr_ref = match CStr::from_ptr((*copts).address).to_str() {
            Ok(v) => v,
            Err(e) => {
                ffi::tarantool::box_error_set(file!().as_ptr() as *const c_char, line!(), TarantoolErrorCode::Unknown as u32, e.to_string().as_ptr() as *const c_char);
                return ptr::null_mut()
            }
        };

        let opts = Options{
            address: addr_ref.to_owned(),
        };
        match App::new(opts) {
            // caller is responsible for the memory
            Ok(v) => Box::into_raw(Box::new(v)),
            Err(e) => {
                ffi::tarantool::box_error_set(file!().as_ptr() as *const c_char, line!(), TarantoolErrorCode::Unknown as u32, e.to_string().as_ptr() as *const c_char);
                return ptr::null_mut()
            }
        }
    }

    #[no_mangle]
    unsafe extern "C" fn ffi_pubsub_stop(app_ptr: *mut App) -> () {
        // rust is responsible for the memory
        let _app = Box::from_raw(app_ptr);
    }

    #[no_mangle]
    unsafe extern "C" fn ffi_pubsub_publish(app_ptr: *mut App, channel: *const c_char, data: *const u8, data_len: usize) -> c_int {
        let channel_ref = match CStr::from_ptr(channel).to_str() {
            Ok(v) => v,
            Err(e) => {
                return ffi::tarantool::box_error_set(file!().as_ptr() as *const c_char, line!(), TarantoolErrorCode::Unknown as u32, e.to_string().as_ptr() as *const c_char);
            }
        };

        let msg = Publication{
            channel: channel_ref.to_owned(),
            data: std::slice::from_raw_parts(data, data_len).to_vec(),
        };
        if let Err(e) = (*app_ptr).send(msg) {
            return ffi::tarantool::box_error_set(file!().as_ptr() as *const c_char, line!(), TarantoolErrorCode::Unknown as u32, e.to_string().as_ptr() as *const c_char);
        }
        0
    }
}
