// OneClick viewer flow: peek metadata, confirm with the user, consume,
// decrypt locally, and redirect to the destination URL.

(function () {
  'use strict';

  const $ = (sel) => document.querySelector(sel);

  const els = {
    loading: $('#loading'),
    confirm: $('#confirm'),
    remaining: $('#remaining'),
    expiry: $('#expiry'),
    btnOpen: $('#btn-open'),
    redirecting: $('#redirecting'),
    destPreview: $('#dest-preview'),
    error: $('#error'),
    errTitle: $('#err-title'),
    errDetail: $('#err-detail'),
  };

  function show(node) {
    for (const c of [els.loading, els.confirm, els.redirecting, els.error]) {
      c.hidden = c !== node;
    }
  }

  function fail(title, detail) {
    els.errTitle.textContent = title;
    els.errDetail.textContent = detail;
    show(els.error);
  }

  function getId() {
    const node = document.getElementById('bootstrap');
    if (!node) return null;
    try {
      return JSON.parse(node.textContent).id;
    } catch (_) {
      return null;
    }
  }

  function getKey() {
    const frag = location.hash.replace(/^#/, '');
    const params = new URLSearchParams(frag);
    const k = params.get('k');
    if (!k) return null;
    try {
      return OneClickCrypto.urlSafeB64ToBytes(k);
    } catch (_) {
      return null;
    }
  }

  async function init() {
    const id = getId();
    if (!id) return fail('Invalid link', 'Missing id.');

    const key = getKey();
    if (!key) return fail('Invalid link', 'The URL is missing the decryption key fragment.');

    let res;
    try {
      res = await fetch('/api/links/' + encodeURIComponent(id) + '/meta');
    } catch (err) {
      return fail('Network error', err && err.message ? err.message : String(err));
    }
    if (!res.ok) {
      return fail(
        'Link unavailable',
        'This link has already been opened, expired, or never existed. By design, we cannot recover it.',
      );
    }
    const meta = await res.json();
    els.remaining.textContent = String(meta.clicks_remaining);
    els.expiry.textContent = 'Expires ' + new Date(meta.expires_at * 1000).toLocaleString();
    show(els.confirm);
  }

  async function reveal() {
    const id = getId();
    const key = getKey();
    if (!id || !key) return fail('Invalid link', 'Missing id or key.');

    show(els.redirecting);
    let res;
    try {
      res = await fetch('/api/links/' + encodeURIComponent(id));
    } catch (err) {
      return fail('Network error', err && err.message ? err.message : String(err));
    }
    if (!res.ok) {
      return fail('Link unavailable', 'Already opened or expired.');
    }
    const body = await res.json();
    const plaintext = await OneClickCrypto.ocDecrypt(body.ciphertext, body.iv, key);
    if (!plaintext) {
      return fail(
        'Could not decrypt',
        "Wrong key in the URL, or the ciphertext was tampered with. Make sure you copied the entire URL, including the part after the '#'.",
      );
    }
    if (!/^https?:\/\//i.test(plaintext)) {
      return fail('Suspicious destination', 'The decrypted value is not an http(s) URL.');
    }
    els.destPreview.textContent = 'Going to ' + plaintext;
    // Small pause so the user can see what they're being redirected to —
    // anti-phishing courtesy. 600ms feels right.
    setTimeout(() => {
      location.replace(plaintext);
    }, 600);
  }

  els.btnOpen.addEventListener('click', reveal);
  init();
})();
