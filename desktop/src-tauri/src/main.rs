#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

use tauri::api::process::{Command, CommandEvent};
use tauri::Manager;

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            // Spawn the Go binary sidecar
            // The sidecar name must match the `externalBin` array in tauri.conf.json
            let (mut rx, child) = Command::new_sidecar("bob-gemini-free")
                .expect("failed to create `bob-gemini-free` binary command")
                .args(["--port", "9610", "--headless"])
                .spawn()
                .expect("Failed to spawn sidecar");

            // Store the child process ID so we can kill it when the app exits if needed
            // Tauri sidecars are automatically killed when the parent process exits,
            // but we can listen to its logs.
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
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
