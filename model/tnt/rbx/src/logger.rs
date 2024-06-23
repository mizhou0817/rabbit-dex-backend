use std::sync::atomic::{AtomicBool, Ordering};

static IS_INITED: AtomicBool = AtomicBool::new(false);

mod ffi {
    use super::*;

    #[no_mangle]
    unsafe extern "C" fn ffi_logger_init() -> () {
        if !IS_INITED.swap(true, Ordering::Relaxed) {
            json_env_logger::init();
            json_env_logger::panic_hook();
        }
    }
}
