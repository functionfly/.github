/**
 * HMAC Request Signer
 * Signs requests with HMAC-SHA256 for enhanced security
 */

import CryptoJS from 'crypto-js';
import type { HMACSignature } from '@/types';

export class HMACRequestSigner {
  private sharedSecret: string;

  constructor(sharedSecret: string) {
    this.sharedSecret = sharedSecret;
  }

  /**
   * Sign request with HMAC-SHA256
   * Signature = HMAC-SHA256(sharedSecret, timestamp + method + path + bodyHash)
   */
  sign(
    method: string,
    path: string,
    body: string = ''
  ): HMACSignature {
    const timestamp = Math.floor(Date.now() / 1000).toString();
    const bodyHash = CryptoJS.SHA256(body).toString();
    const signatureString = `${timestamp}${method}${path}${bodyHash}`;

    const signature = CryptoJS.HmacSHA256(
      signatureString,
      this.sharedSecret
    ).toString();

    return { timestamp, signature };
  }

  /**
   * Verify HMAC signature (for testing)
   */
  verify(
    method: string,
    path: string,
    body: string,
    timestamp: string,
    signature: string
  ): boolean {
    const bodyHash = CryptoJS.SHA256(body).toString();
    const signatureString = `${timestamp}${method}${path}${bodyHash}`;
    const expectedSignature = CryptoJS.HmacSHA256(
      signatureString,
      this.sharedSecret
    ).toString();
    return signature === expectedSignature;
  }
}
