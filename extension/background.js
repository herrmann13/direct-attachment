// Enables Direct Attachment only in the tab explicitly activated by the user.
// `activeTab` grants temporary access to that tab after a toolbar-icon click.
(() => {
  "use strict";

  const api = globalThis.browser || globalThis.chrome;
  const defaultTitle = "Ativar Direct Attachment nesta aba";
  const enabledTitle = "Direct Attachment está ativo nesta aba";
  const injectionFiles = [
    "lib/qrcode.min.js",
    "lib/tweetnacl.min.js",
    "content/browser.js",
    "content/styles.js",
    "content/crypto.js",
    "content/config.js",
    "content/interceptor.js",
    "content/overlay.js",
    "content/transfer.js",
    "content/injector.js",
    "content/index.js",
  ];
  const injectingTabs = new Set();

  function setActionState(tabId, active) {
    api.action.setBadgeText({ tabId, text: active ? "ON" : "" });
    api.action.setTitle({ tabId, title: active ? enabledTitle : defaultTitle });
  }

  async function isAlreadyActive(tabId) {
    const result = await api.scripting.executeScript({
      target: { tabId },
      func: () => Boolean(globalThis.DirectAttachment?.interceptor?.isActive),
    });
    return Boolean(result[0]?.result);
  }

  async function activateTab(tab) {
    if (!tab.id || injectingTabs.has(tab.id)) return;

    injectingTabs.add(tab.id);

    try {
      if (await isAlreadyActive(tab.id)) {
        setActionState(tab.id, true);
        return;
      }

      await api.scripting.executeScript({
        target: { tabId: tab.id },
        files: injectionFiles,
      });
      setActionState(tab.id, true);
    } catch (error) {
      // Browser-internal pages and the Chrome Web Store do not allow injection.
      console.warn("Direct Attachment could not be activated in this tab.", error);
      api.action.setBadgeText({ tabId: tab.id, text: "!" });
      api.action.setTitle({ tabId: tab.id, title: "Não é possível ativar nesta página" });
    } finally {
      injectingTabs.delete(tab.id);
    }
  }

  api.action.onClicked.addListener((tab) => {
    void activateTab(tab);
  });

  // A navigation replaces the document, so injected listeners and activeTab
  // access no longer apply. The user must explicitly activate the new page.
  api.tabs.onUpdated.addListener((tabId, changeInfo) => {
    if (changeInfo.status === "loading") setActionState(tabId, false);
  });

  api.tabs.onRemoved.addListener((tabId) => {
    injectingTabs.delete(tabId);
  });
})();
