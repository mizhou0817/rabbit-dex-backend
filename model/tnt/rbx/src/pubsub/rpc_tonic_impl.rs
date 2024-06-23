use super::Publication;
use super::centrifugo::{
    centrifugo_api_client::CentrifugoApiClient,
    Command,
    command::MethodType,
    PublishRequest,
};

use anyhow::bail;
use tonic::transport::{Channel, Uri};

pub(crate) type BatchRequest = super::centrifugo::BatchRequest;

type Result<T> = std::result::Result<T, anyhow::Error>;

pub(crate) struct CentrifugoClient {
    inner: CentrifugoApiClient<Channel>,
}

impl CentrifugoClient {
    pub async fn new(address: &str) -> Result<Self> {
        let uri = address.parse::<Uri>()?;

        let channel = Channel::builder(uri).connect_lazy();
        Ok(Self{
            inner: CentrifugoApiClient::new(channel)
        })
    }

    pub async fn publish(&mut self, batch: Vec<Publication>) -> Result<()> {
        let mut req = BatchRequest{commands: Vec::new()};
        let mut index: u32 = 0;
        for msg in batch {
            index += 1;
            req.commands.push(Command{
                id: index,
                method: MethodType::Publish as i32,
                params: Vec::new(),
                publish: Some(PublishRequest{
                    channel: msg.channel,
                    data: msg.data,
                    b64data: String::new(),
                    skip_history: false,
                    tags: std::collections::HashMap::new(),
                }),
                broadcast: None,
                subscribe: None,
                unsubscribe: None,
                disconnect: None,
                presence: None,
                presence_stats: None,
                history: None,
                history_remove: None,
                info: None,
                rpc: None,
                refresh: None,
                channels: None,
                connections: None,
                update_user_status: None,
                get_user_status: None,
                delete_user_status: None,
                block_user: None,
                unblock_user: None,
                revoke_token: None,
                invalidate_user_tokens: None,
                device_register: None,
                device_update: None,
                device_remove: None,
                device_list: None,
                device_topic_list: None,
                device_topic_update: None,
                user_topic_list: None,
                user_topic_update: None,
                send_push_notification: None,
                update_push_status: None,
            });
        }
        match self.inner.batch(req).await {
            Err(e) => bail!("transport error: {}", e.message()),
            Ok(v) => for reply in v.into_inner().replies {
                if let Some(err) = reply.error {
                    bail!("logic error: {}, reply-id:{}", err.message, reply.id)
                }
            },
        }

        Ok(())
    }
}
