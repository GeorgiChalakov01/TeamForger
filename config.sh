#!/bin/bash

# DB
DB_CONTAINER_NAME="teamforger-db-1"
DB_PWD="ChangeMe" # Don't add spaces and be careful with special symbols
DB_SCHEMA="TF"
DB_USER="user"
DB_PORT="5432"

# BE
BE_HOST="teamforger.gchalakov.com"
BE_PORT="8080"
ALLOWED_WS_ORIGIN="https://teamforger.gchalakov.com"
OLLAMA_API="http://192.168.0.27:11434/api/chat"
OLLAMA_MODEL="gemma3:12b" #"gemma3:4b-it-qat" #"qwen3:4b" #"hf.co/Qwen/Qwen3-8B-GGUF:Q8_0"
OLLAMA_CTX="4096"
OLLAMA_EMB_API="http://192.168.0.27:11434/api/embeddings"
OLLAMA_EMB_MODEL="nomic-embed-text"
