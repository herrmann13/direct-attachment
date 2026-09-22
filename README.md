# Direct Attachment

Anexe fotos do seu celular a qualquer site pelo Chrome ou Firefox usando um
QR Code — sem instalar nada no celular.

Quando você clica em "Anexar arquivo" em um site, a extensão intercepta o clique,
mostra um QR Code. Você escaneia com o celular, tira/usa uma foto, e o arquivo
aparece no `<input type="file">` como se tivesse sido selecionado normalmente.

> **Compatibilidade:** a mesma extensão funciona em Chrome e Firefox (MV3). A
> diferença está no último passo: o **Chrome** injeta o arquivo no input
> automaticamente; o **Firefox** não permite isso (o `input.files` é somente
> leitura), então nele o arquivo é entregue por *drop* sintético e, se
> necessário, por download manual. Veja [Suporte a navegadores](#suporte-a-navegadores).

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
- **`extension/`** — Extensão WebExtension MV3 (Chrome e Firefox). Intercepta o
  clique, mostra o QR e injeta o arquivo recebido no input original.
- A foto é **cifrada de ponta a ponta** (TweetNaCl/XSalsa20-Poly1305): o servidor
  só repassa texto cifrado e **nunca consegue ver a imagem**. A chave viaja
  apenas no fragmento do QR Code (`#k=...`), que o navegador não envia ao servidor.

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

## Extensão (Chrome e Firefox)

### 1. Configurar o backend

Edite `extension/content/config.js` e aponte `backendOrigin` para o seu backend:

```js
DirectAttachment.config = {
  backendOrigin:"https://direct-attachment-production.up.railway.app",
};
```

Para desenvolvimento local, use `http://localhost:8080` (o Chrome trata
`localhost` como contexto seguro).

### 2. Definir o ID do Firefox

Edite `extension/manifest.json` e troque o `id` em `browser_specific_settings`
por um identificador seu (ex.: `direct-attachment@seudominio.com`). Esse ID é
permanente na AMO e é usado pelo `update.json`.

### 3. Carregar a extensão (desenvolvimento)

**Chrome**

1. Abra `chrome://extensions`.
2. Ative o **Modo do desenvolvedor**.
3. Clique em **Carregar sem compactação** e selecione a pasta `extension/`.

**Firefox**

1. Abra `about:debugging#/runtime/this-firefox`.
2. Clique em **Carregar complemento temporário**.
3. Selecione o `manifest.json` dentro de `extension/`.

### 4. Usar

1. Em qualquer site, clique no botão de anexar arquivo.
2. No overlay, escaneie o QR Code com o celular.
3. Tire uma foto ou escolha uma da galeria.
4. O arquivo aparece no input como se tivesse sido selecionado normalmente
   (Chrome) ou é entregue via drop/download (Firefox).

## Suporte a navegadores

| Navegador | Anexar automaticamente no input | Comportamento |
|-----------|---------------------------------|---------------|
| Chrome    | ✅ Sim                          | Injeção via `DataTransfer` (o arquivo aparece no input como se tivesse sido escolhido). |
| Firefox   | ❌ Não                          | `input.files` é somente leitura. A extensão tenta um `drop` sintético (funciona em sites com drag-and-drop) e oferece um botão **Baixar arquivo** como fallback. |

## Publicar / distribuir

### Empacotar

```bash
./scripts/package.sh
```

Gera em `dist/` os pacotes `direct-attachment-chrome-<versão>.zip`,
`direct-attachment-firefox-<versão>.zip` e o `update.json` (template de
auto-atualização do Firefox).

### Distribuição própria (sem loja)

**Firefox** — a distribuição própria é suportada nativamente:

1. Em [addons.mozilla.org](https://addons.mozilla.org), envie o
   `direct-attachment-firefox-<versão>.zip` como **unlisted** (só para assinatura).
2. Baixe o `.xpi` assinado que a AMO devolve.
3. Hospede o `.xpi` (GitHub Releases, Railway, etc.) e edite o `update_link` no
   `dist/update.json` para apontar para ele.
4. Sirva o `update.json` em HTTPS; o Firefox usará essa URL para auto-atualizar.

**Chrome** — o Chrome **bloqueia instalação fora da Web Store** (exceto Modo do
desenvolvedor ou política enterprise). Para distribuição própria:

- Entregue o `.zip` + instruções de "Carregar sem compactação" (dev mode) para
  usuários de teste; ou
- Publique como **Unlisted** na Chrome Web Store (link-only, ainda exige conta
  de desenvolvedor de US$ 5 e revisão).

### Publicação pública (quando validado)

- **Chrome Web Store:** conta de desenvolvedor (US$ 5, taxa única) → enviar o
  `.zip` → preencher listagem/privacy → justificar `host_permissions: <all_urls>`
  e o uso do serviço externo (backend) → submeter para revisão.
- **Firefox AMO:** conta gratuita → enviar o `.zip` como **listado** → revisão.

Em ambos, descreva com precisão os dados necessários à transferência: a imagem
é **cifrada de ponta a ponta**, portanto apenas o texto cifrado transita pelo
servidor, que não tem acesso ao conteúdo da foto. O nome/tipo do arquivo e
identificadores temporários de sessão também são transmitidos para concluir o
envio.

> Para a Chrome Web Store, não declare que a extensão não coleta dados: a foto
> cifrada, o nome/tipo do arquivo e identificadores temporários de sessão são
> transmitidos para executar a transferência. A política pública já está em
> `https://direct-attachment-production.up.railway.app/privacy.html`; os textos
> para a ficha, a justificativa de `"<all_urls>"` e a revisão estão em
> [`CHROME_WEB_STORE.md`](CHROME_WEB_STORE.md).

## Privacidade e criptografia

- A foto é cifrada no celular com **TweetNaCl** (`nacl.secretbox`,
  XSalsa20-Poly1305) usando uma chave efêmera de 256 bits gerada no PC.
- A chave viaja somente no fragmento do QR Code (`#k=...`); fragmentos não são
  enviados ao servidor em requisições HTTP, então o servidor nunca conhece a chave.
- O servidor repassa apenas os bytes cifrados (opacos) e não os inspeciona nem
  armazena; a sessão é apagada logo após a entrega.
- Validação de tipo/tamanho: no celular (antes de cifrar) e no PC (após decifrar).

## Limitações conhecidas

- Sites que abrem o seletor via `input.showPicker()` (em vez de `.click()`) não
  disparam um evento `click`, portanto não são interceptados.
- No Firefox, o arquivo não é anexado automaticamente ao input (limitação da
  plataforma); usa drop sintético + download como fallback.
- A câmera do celular exige contexto seguro (HTTPS). O Railway já fornece HTTPS;
  em desenvolvimento local use `localhost`.

## Estrutura de diretórios

```
server/
  cmd/server/main.go         # wiring + shutdown gracioso
  internal/config/           # env vars
  internal/session/          # domínio + store em memória
  internal/httpx/            # respostas/erros JSON padronizados
  internal/upload/           # metadados do arquivo (nome), sem sniff
  internal/transfer/         # hub WebSocket (relay de bytes opacos)
  internal/api/              # rotas, handlers e middleware
  web/                       # página do celular (embed) + tweetnacl
extension/
  manifest.json
  content/                   # content scripts (browser, crypto, config,
                             #   interceptor, overlay, transfer, injector, index)
  overlay.css
  icons/                     # 48/96/128 px
  lib/qrcode.min.js          # biblioteca de QR (vendored)
  lib/tweetnacl.min.js       # criptografia E2E (vendored)
scripts/
  package.sh                 # gera os pacotes de distribuição
```
