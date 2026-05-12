// OneClick crypto — Web Crypto API helpers.
//
// AES-GCM, 256-bit key, 96-bit IV. The 32-byte key is generated in the
// browser, exported as raw bytes, and embedded in the URL fragment so the
// server never sees it.

const KEY_BYTES = 32;
const IV_BYTES = 12;

async function ocEncrypt(plaintext) {
  const enc = new TextEncoder().encode(plaintext);
  const key = crypto.getRandomValues(new Uint8Array(KEY_BYTES));
  const iv = crypto.getRandomValues(new Uint8Array(IV_BYTES));
  const cryptoKey = await crypto.subtle.importKey('raw', key, 'AES-GCM', false, ['encrypt']);
  const ct = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, cryptoKey, enc);
  return { ciphertext: bytesToB64(new Uint8Array(ct)), iv: bytesToB64(iv), key };
}

async function ocDecrypt(ciphertextB64, ivB64, rawKey) {
  try {
    const ct = b64ToBytes(ciphertextB64);
    const iv = b64ToBytes(ivB64);
    if (rawKey.length !== KEY_BYTES) return null;
    if (iv.length !== IV_BYTES) return null;
    const cryptoKey = await crypto.subtle.importKey('raw', rawKey, 'AES-GCM', false, ['decrypt']);
    const out = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, cryptoKey, ct);
    return new TextDecoder().decode(out);
  } catch (_) {
    return null;
  }
}

function bytesToB64(bytes) {
  let bin = '';
  for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i]);
  return btoa(bin);
}

function b64ToBytes(s) {
  const bin = atob(s);
  const out = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
  return out;
}

function bytesToUrlSafeB64(bytes) {
  return bytesToB64(bytes).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function urlSafeB64ToBytes(s) {
  let std = s.replace(/-/g, '+').replace(/_/g, '/');
  const pad = (4 - (std.length % 4)) % 4;
  std += '='.repeat(pad);
  return b64ToBytes(std);
}

window.OneClickCrypto = { ocEncrypt, ocDecrypt, bytesToUrlSafeB64, urlSafeB64ToBytes };
