#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

use std::sync::Mutex;

use tauri::api::process::{Command, CommandChild, CommandEvent};
use tauri::{Manager, RunEvent};

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            // Spawn the Go binary sidecar
            // The sidecar name must match the `externalBin` array in tauri.conf.json
            let (mut rx, child) = Command::new_sidecar("bob-gemini-free")
                .expect("failed to create `bob-gemini-free` binary command")
                .args(["--port", "9610", "--headless", "--config", "none"])
                .spawn()
                .expect("Failed to spawn sidecar");

            // Keep the child owned by the application. Tauri also terminates
            // sidecars during app shutdown, while this managed state makes the
            // lifecycle ownership explicit instead of dropping the handle.
            app.manage(Mutex::new(Some(child)));

            tauri::async_runtime::spawn(async move {
                while let Some(event) = rx.recv().await {
                    if let CommandEvent::Stdout(line) = event {
                        println!("Engine: {}", line);
                    } else if let CommandEvent::Stderr(line) = event {
                        eprintln!("Engine Error: {}", line);
                    }
                }
            });

            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|app, event| {
            if matches!(event, RunEvent::ExitRequested { .. } | RunEvent::Exit) {
                if let Some(child) = app
                    .state::<Mutex<Option<CommandChild>>>()
                    .lock()
                    .expect("sidecar state mutex was poisoned")
                    .take()
                {
                    if let Err(error) = child.kill() {
                        eprintln!("Failed to stop gateway sidecar: {error}");
                    }
                }
            }
        });
}
