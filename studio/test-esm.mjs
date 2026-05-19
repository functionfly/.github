import { createRequire } from 'module';
const require = createRequire(import.meta.url);
const electron = require('electron');
console.log('electron type:', typeof electron);
console.log('electron keys:', Object.keys(electron).slice(0,5));
console.log('app:', typeof electron.app);
