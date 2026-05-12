// OneClick compose page logic — wires the form to the crypto helpers
// and POSTs the resulting ciphertext to /api/links.

(function () {
  'use strict';

  const $ = (sel) => document.querySelector(sel);

  const els = {
    urlInput: $('#url-input'),
    ttl: $('#ttl'),
    clicks: $('#clicks'),
    btnCreate: $('#btn-create'),
    err: $('#err'),
    composeCard: $('#compose-card'),
    resultCard: $('#result-card'),
    resultUrl: $('#result-url'),
    btnCopy: $('#btn-copy'),
    btnNew: $('#btn-new'),
    meta: $('#meta'),
  };

  function showError(msg) {
    els.err.textContent = msg;
    els.err.hidden = false;
  }

  function clearError() {
    els.err.hidden = true;
  }

  async function submit() {
    clearError();
    const value = (els.urlInput.value || '').trim();
    if (!value) {
      showError('Type a URL first.');
      return;
    }
    if (!/^https?:\/\//i.test(value)) {
      showError('URL must start with http:// or https://.');
      return;
    }
    if (value.length > 2000) {
      showError('URL is too long (max 2000 characters).');
      return;
    }

    els.btnCreate.disabled = true;
    els.btnCreate.textContent = 'Encrypting…';

    let env;
    try {
      env = await OneClickCrypto.ocEncrypt(value);
    } catch (err) {
      els.btnCreate.disabled = false;
      els.btnCreate.textContent = 'Encrypt & shorten';
      showError('Encryption failed: ' + (err && err.message ? err.message : String(err)));
      return;
    }

    let res;
    try {
      res = await fetch('/api/links', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ciphertext: env.ciphertext,
          iv: env.iv,
          ttl_seconds: Number(els.ttl.value),
          max_clicks: Number(els.clicks.value),
        }),
      });
    } catch (err) {
      els.btnCreate.disabled = false;
      els.btnCreate.textContent = 'Encrypt & shorten';
      showError('Network error: ' + (err && err.message ? err.message : String(err)));
      return;
    }

    els.btnCreate.disabled = false;
    els.btnCreate.textContent = 'Encrypt & shorten';

    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      showError('Server rejected the link: ' + (body.error || res.statusText));
      return;
    }

    const body = await res.json();
    const fragment = OneClickCrypto.bytesToUrlSafeB64(env.key);
    const url = location.origin + '/l/' + body.id + '#k=' + fragment;
    els.resultUrl.value = url;
    const expiry = new Date(body.expires_at * 1000);
    els.meta.textContent =
      'Expires ' + expiry.toLocaleString() +
      ' • burns after ' + body.clicks_remaining +
      ' click' + (body.clicks_remaining === 1 ? '' : 's') + '.';

    els.composeCard.hidden = true;
    els.resultCard.hidden = false;
    els.resultUrl.focus();
    els.resultUrl.select();
    els.urlInput.value = '';
  }

  async function copy() {
    try {
      await navigator.clipboard.writeText(els.resultUrl.value);
      els.btnCopy.textContent = 'Copied ✓';
      setTimeout(() => (els.btnCopy.textContent = 'Copy'), 1500);
    } catch (err) {
      console.warn('clipboard.writeText failed:', err);
      els.resultUrl.focus();
      els.resultUrl.select();
    }
  }

  function reset() {
    els.composeCard.hidden = false;
    els.resultCard.hidden = true;
    els.urlInput.focus();
  }

  els.btnCreate.addEventListener('click', submit);
  els.btnCopy.addEventListener('click', copy);
  els.btnNew.addEventListener('click', reset);
  els.urlInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') submit();
  });

  els.urlInput.focus();
})();
