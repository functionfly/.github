/**
 * Analytics module — single entry point for all product analytics.
 *
 * Exports:
 *   trackEvent(name, payload)  – track a product event
 *   trackPageView(path)        – track a page navigation
 *   identifyUser(id, traits)    – identify the logged-in user
 *   resetUser()                – clear identity on logout
 *
 * Implementation: Mixpanel Browser SDK (zero-knowledge, client-side only).
 * Consent-gated via the cookie-consent "analytics" category.
 */
export { identifyUser, resetUser, trackEvent, trackPageView } from './mixpanel';
