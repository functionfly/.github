import { createRequire } from 'module';
const require = createRequire(import.meta.url);
const e = require('electron');
console.log('Keys:', Object.keys(e).slice(0,5));
console.log('app type:', typeof e.app);
console.log('app:', e.app);
