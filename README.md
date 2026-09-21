# Direct Attachment

Anexe fotos do seu celular a qualquer site pelo Chrome usando um QR Code —
sem instalar nada no celular.

Quando você clica em "Anexar arquivo" em um site, a extensão intercepta o clique,
mostra um QR Code. Você escaneia com o celular, tira/usa uma foto, e o arquivo
aparece no `<input type="file">` como se tivesse sido selecionado normalmente.

## Arquitetura

```
Site (PC)  ──clique interceptado──▶  Extensão (content script)
   input[type=file] ◀──File injetado──  overlay + QR + WebSocket
                                              │ wss
                                     Backend Go (Railway)
                                              ▲ POST multipart
                                     Celular (página web)
```

- **`server/`** — Backend Go. Cria sessões de pareamento, serve a página do
  celular e faz o *relay* do arquivo via WebSocket.
- **`extension/`** — Extensão Chrome MV3. Intercepta o clique, mostra o QR e
  injeta o arquivo recebido no input original.
- A foto **passa pelo servidor** (relay), o que é simples e confiável.

## Backend (Go)

### Variáveis de ambiente

| Variável       | Padrão       | Descrição                                        |
|----------------|--------------|--------------------------------------------------|
| `PORT`         | `8080`       | Porta HTTP.                                      |
| `BASE_URL`     | *(derivado)* | Origem pública usada no QR. Se vazia, deriva da requisição (`Host` + `X-Forwarded-Proto`). |
| `SESSION_TTL`  | `2m`         | Tempo de vida de uma sessão (ex.: `2m`, `90s`).  |
| `MAX_FILE_SIZE`| `15728640`   | Tamanho máximo do upload em bytes (15 MiB).      |

### Rodando localmente

```bash
cd server
go run ./cmd/server
# http://localhost:8080
```

Testes:

```bash
cd server
go test ./...
```

### Deploy no Railway

1. Crie um novo projeto no Railway apontando para este repositório.
2. O Railway detecta o `railway.toml` (Dockerfile em `server/`).
3. Defina `PORT=8080` (o Railway injeta automaticamente; o binário escuta em
   `:$PORT`). Opcionalmente defina `BASE_URL` com o domínio público gerado.
4. Copie o domínio público (ex.: `https://direct-attachment.up.railway.app`).

> **Nota:** o QR Code aponta para `BASE_URL`. Se você não definir `BASE_URL`,
> o servidor deriva a URL da própria requisição, então normalmente funciona
> sem configuração extra atrás do proxy do Railway.

## Extensão (Chrome)

### 1. Configurar o backend

Edite `extension/content/config.js` e aponte `backendOrigin` para o seu backend:

```js
DirectAttachment.config = {
  backendOrigin: "https://direct-attachment.up.railway.app",
};
```

Para desenvolvimento local, use `http://localhost:8080` (o Chrome trata
`localhost` como contexto seguro).

### 2. Carregar a extensão

1. Abra `chrome://extensions`.
2. Ative o **Modo do desenvolvedor**.
3. Clique em **Carregar sem compactação** e selecione a pasta `extension/`.

### 3. Usar

1. Em qualquer site, clique no botão de anexar arquivo.
2. No overlay, escaneie o QR Code com o celular.
3. Tire uma foto ou escolha uma da galeria.
4. O arquivo aparece no input como se tivesse sido selecionado normalmente.

## Limitações conhecidas

- Sites que abrem o seletor via `input.showPicker()` (em vez de `.click()`) não
  disparam um evento `click`, portanto não são interceptados.
- A câmera do celular exige contexto seguro (HTTPS). O Railway já fornece HTTPS;
  em desenvolvimento local use `localhost`.

## Estrutura de diretórios

```
server/
  cmd/server/main.go         # wiring + shutdown gracioso
  internal/config/           # env vars
  internal/session/          # domínio + store em memória
  internal/httpx/            # respostas/erros JSON padronizados
  internal/upload/           # validação de tipo/tamanho
  internal/transfer/         # hub WebSocket (relay)
  internal/api/              # rotas, handlers e middleware
  web/                       # página do celular (embed)
extension/
  manifest.json
  content/                   # content scripts (config, interceptor, overlay,
                             #   transfer, injector, index)
  overlay.css
  lib/qrcode.min.js          # biblioteca de QR (vendored)
```
