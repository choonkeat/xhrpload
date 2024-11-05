package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
)

type Flags struct {
	listenAddr string
	outputPath string
}

var uploadedBytes int64

func main() {
	var flags Flags
	// Define the output flag
	flag.StringVar(&flags.listenAddr, "listen", ":8080", "Host and port to listen")
	flag.StringVar(&flags.outputPath, "output", filepath.Join(os.Getenv("HOME"), "Downloads"), "File to write uploaded content")
	flag.Parse()

	// Handle root route to serve HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		htmlContent := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>File Upload</title>
			<style>
				.progress-bar {
					width: 100%;
					background-color: #f3f3f3;
					margin-bottom: 10px;
				}
				.progress-bar div {
					width: 0%;
					height: 20px;
					background-color: #2196F3;
					text-align: center;
					line-height: 20px;
					color: white;
				}
				.completed {
					background-color: green !important;
				}
			</style>
		</head>
		<body>
			<h1>Upload Files</h1>
			<input type="file" id="fileInput" multiple><br><br>
			<div id="progressContainer"></div>
			<script>
				document.getElementById('fileInput').addEventListener('change', function() {
					var files = this.files;
					for (var i = 0; i < files.length; i++) {
						uploadFile(files[i]);
					}
				});
				
				function uploadFile(file) {
					var progressContainer = document.getElementById('progressContainer');

					var progressDiv = document.createElement('div');

					var label = document.createElement('div');
					label.textContent = file.name;
					progressDiv.appendChild(label);

					var progressBar = document.createElement('div');
					progressBar.className = 'progress-bar';
					var progressBarFill = document.createElement('div');
					progressBar.appendChild(progressBarFill);
					progressDiv.appendChild(progressBar);

					progressContainer.appendChild(progressDiv);

					var xhr = new XMLHttpRequest();
					xhr.open('POST', '/upload?filename=' + encodeURIComponent(file.name), true);
					xhr.setRequestHeader('Content-Type', 'application/octet-stream');

					xhr.upload.onprogress = function(event) {
						if (event.lengthComputable) {
							var percentComplete = (event.loaded / event.total) * 100;
							progressBarFill.style.width = percentComplete + '%';
							progressBarFill.textContent = percentComplete.toFixed(2) + '%';
						}
					};

					xhr.onload = function() {
						if (xhr.status === 200) {
							progressBarFill.classList.add('completed');
							progressBarFill.textContent = 'Completed';
						} else {
							progressBarFill.style.backgroundColor = 'red';
							progressBarFill.textContent = 'Failed';
						}
					};

					xhr.send(file);
				}
			</script>
		</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlContent)
	})

	// Handle file upload
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		atomic.StoreInt64(&uploadedBytes, 0)

		log.Printf("Received request from %s", r.RemoteAddr)
		defer func() {
			log.Printf("Request from %s completed (uploaded %d bytes)", r.RemoteAddr, uploadedBytes)
		}()

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		filename := filepath.Base(r.URL.Query().Get("filename"))
		outputFilePath := filepath.Join(flags.outputPath, filename)
		file, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			http.Error(w, "Unable to open output file", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Error closing file: %v", err)
			}
		}()

		bufferSize := 8192
		buffer := make([]byte, bufferSize)
		for {
			n, err := r.Body.Read(buffer)
			if n > 0 {
				_, writeErr := file.Write(buffer[:n])
				if writeErr != nil {
					http.Error(w, "Unable to write to output file", http.StatusInternalServerError)
					return
				}
				atomic.AddInt64(&uploadedBytes, int64(n))
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "File uploaded successfully")
	})

	// Start the server
	log.Printf("Listening on %s", flags.listenAddr)
	if err := http.ListenAndServe(flags.listenAddr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
