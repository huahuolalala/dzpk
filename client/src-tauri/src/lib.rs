use futures_util::{SinkExt, StreamExt, TryStreamExt};
use log::{error, info};
use std::sync::Arc;
use tauri::{AppHandle, Emitter, State};
use tokio::sync::{mpsc, RwLock};
use tokio_tungstenite::{connect_async, tungstenite::Message};

#[derive(Default)]
pub struct WsState {
    pub sender: Arc<RwLock<Option<mpsc::Sender<String>>>>,
}

#[tauri::command]
async fn ws_connect(url: String, state: State<'_, WsState>, app: AppHandle) -> Result<(), String> {
    let (tx, mut rx) = mpsc::channel::<String>(100);

    {
        let mut sender = state.sender.write().await;
        *sender = Some(tx.clone());
    }

    let app_handle = app.clone();
    let sender_clone = state.sender.clone();

    tokio::spawn(async move {
        match connect_async(&url).await {
            Ok((ws_stream, _)) => {
                info!("WebSocket connected to {}", url);
                let (mut write, mut read) = ws_stream.split();

                // 发送消息任务
                tokio::spawn(async move {
                    while let Some(msg) = rx.recv().await {
                        if let Err(e) = write
                            .send(Message::Text(msg.into()))
                            .await
                        {
                            error!("Send error: {}", e);
                            break;
                        }
                    }
                });

                // 接收消息任务
                let app_handle_clone = app_handle.clone();
                tokio::spawn(async move {
                    while let Ok(msg) = read.try_next().await {
                        match msg {
                            Some(Message::Text(text)) => {
                                let _ = app_handle_clone.emit("ws_message", text.to_string());
                            }
                            Some(Message::Close(_)) => {
                                let _ = app_handle_clone.emit("ws_closed", ());
                                break;
                            }
                            None => break,
                            _ => {}
                        }
                    }
                });
            }
            Err(e) => {
                error!("WebSocket connection error: {}", e);
                let _ = app_handle.emit("ws_error", e.to_string());
            }
        }

        // 清理 sender
        let mut sender = sender_clone.write().await;
        *sender = None;
    });

    Ok(())
}

#[tauri::command]
async fn ws_send(message: String, state: State<'_, WsState>) -> Result<(), String> {
    let sender = state.sender.read().await;
    if let Some(tx) = sender.as_ref() {
        tx.send(message).await.map_err(|e| e.to_string())?;
    }
    Ok(())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    env_logger::init();
    tauri::Builder::default()
        .manage(WsState::default())
        .plugin(tauri_plugin_opener::init())
        .invoke_handler(tauri::generate_handler![ws_connect, ws_send])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
