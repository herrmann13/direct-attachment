# Submissão à Chrome Web Store

Este documento contém os textos e a lista de verificação para publicar o Direct Attachment.

## URLs para a ficha da loja

| Campo | Valor |
|---|---|
| Homepage | `https://github.com/herrmann13/direct-attachment` |
| Política de privacidade | `https://direct-attachment-production.up.railway.app/privacy.html` |
| Suporte | `https://github.com/herrmann13/direct-attachment/issues` |

Antes de submeter, abra a URL da política em uma janela anônima. Se o domínio Railway mudar, atualize este arquivo e a URL informada no painel da loja.

## Texto da listagem

**Resumo**

> Anexe fotos do celular a campos de upload usando um QR Code, sem instalar nada no telefone.

**Descrição detalhada**

> Direct Attachment permite transferir uma foto do celular para o campo de anexo que você abriu no navegador. Ao clicar em um campo de upload, a extensão mostra um QR Code. Escaneie-o com o celular, tire ou escolha uma foto e confirme o envio.
>
> A foto é cifrada de ponta a ponta antes de sair do celular. O servidor apenas retransmite bytes cifrados para o navegador que iniciou a sessão e não recebe a chave de decifragem. As sessões são temporárias, de uso único e expiram rapidamente.
>
> Use a extensão somente em sites nos quais você deseja anexar uma imagem. Você pode cancelar a transferência ou escolher usar o seletor nativo de arquivos do computador.

## Justificativa da permissão de sites

Informe, no campo de justificativa de permissões:

> A permissão de acesso aos sites é necessária para que o content script identifique interações com campos `<input type="file">` em qualquer site que o usuário escolher. A função principal da extensão é oferecer, nesse campo, a opção de anexar uma foto enviada pelo celular. O código é executado somente para detectar esse clique e inserir no campo o arquivo que o próprio usuário confirmou; ele não lê páginas para criar perfis, não altera conteúdo para publicidade e não acessa histórico de navegação.

Se o painel separar a justificativa do host externo, informe:

> A extensão se conecta exclusivamente a `https://direct-attachment-production.up.railway.app` para criar sessões temporárias e retransmitir o arquivo cifrado entre o celular e o navegador. Essa conexão é essencial para exibir o QR Code e concluir a transferência solicitada pelo usuário.

## Privacidade: respostas para o painel

Confirme as opções exibidas pelo painel atual, mas use estas respostas como base:

| Pergunta | Resposta |
|---|---|
| A extensão coleta ou transmite dados do usuário? | Sim, somente para executar a transferência iniciada pelo usuário. |
| Quais categorias? | Conteúdo do usuário (foto cifrada), identificadores temporários de sessão, nome/tipo do arquivo e dados técnicos de requisição. |
| A foto é vendida, usada para anúncios ou treinamento de IA? | Não. |
| A foto fica armazenada permanentemente? | Não. Ela é retransmitida em uma sessão de uso único; sessões expiram em até dois minutos. |
| Os dados são criptografados em trânsito? | Sim, a conexão usa HTTPS/WSS e a foto recebe uma camada adicional de criptografia de ponta a ponta antes do upload. |
| O tratamento é necessário para a funcionalidade? | Sim. A transferência não funciona sem esses dados temporários. |

Não marque “nenhum dado é coletado”: o serviço transmite a foto cifrada e seus metadados para realizar a função solicitada.

## Materiais e revisão final

1. Gere o pacote com `./scripts/package.sh` e envie somente `dist/direct-attachment-chrome-<versão>.zip`.
2. Faça capturas mostrando: clique em um campo de anexo, QR Code, página do celular e foto recebida.
3. Teste a extensão empacotada no Chrome, em um site HTTPS com campo de upload, usando o backend de produção.
4. Abra a política de privacidade pela URL pública e informe-a no painel.
5. Revise se a versão em `extension/manifest.json` foi incrementada desde a última publicação.
6. Publique primeiro como **Não listado** para validar a revisão e o fluxo real de instalação.

## Observação sobre o escopo de sites

O manifesto mantém `"<all_urls>"` porque a proposta da extensão é funcionar em qualquer página que possua campo de anexo. Restringir essa permissão a uma lista fixa de domínios impediria o funcionamento em outros sites. A justificativa acima deve ser mantida consistente com a descrição e com a política de privacidade.
