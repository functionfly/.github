/**
 * FunctionFly SDK
 *
 * A JavaScript/TypeScript SDK for compiling deterministic functions to WebAssembly
 * for execution on the FunctionFly platform.
 *
 * Example:
 *     import { function, StateClient } from '@functionfly/flypy';
 *
 *     // Register a deterministic function
 *     function({
 *       name: 'calculate-total',
 *       deterministic: true,
 *       idempotent: true,
 *       cacheTtl: 3600
 *     })
 *     async function handler(event) {
 *       const items = event.items || [];
 *       const taxRate = event.tax_rate || 0.08;
 *
 *       const subtotal = items.reduce((sum, item) => sum + item.price * item.quantity, 0);
 *       const tax = subtotal * taxRate;
 *
 *       return {
 *         subtotal,
 *         tax,
 *         total: subtotal + tax
 *       };
 *     }
 *
 *     // Use StateFabric
 *     const state = new StateClient({ apiKey: 'your-key' });
 *     await state.setValue('tenant/cart/user123', { items: [] });
 */
export * from './types.js';
export * from './state.js';
export { getValue, setValue, deleteValue, getHistory, createSnapshot, restoreSnapshot, } from './state.js';
export declare const VERSION = "1.1.0";
//# sourceMappingURL=index.d.ts.map