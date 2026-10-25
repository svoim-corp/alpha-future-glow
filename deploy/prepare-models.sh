#!/bin/sh
set -eu

# На повторном и автономном запуске не обращаемся к реестру.
if ! (cd /root/.ollama/models && sha256sum -c /config/models.sha256) 2>/dev/null; then
  ollama pull "$CHAT_MODEL"
  ollama pull "$EMBED_MODEL"
  (cd /root/.ollama/models && sha256sum -c /config/models.sha256)
fi
ollama list
