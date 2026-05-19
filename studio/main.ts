import { app, BrowserWindow, ipcMain, shell } from 'electron';
import * as path from 'path';

const isDev = process.env.NODE_ENV === 'development' || !app.isPackaged;

if (app) {
  app.disableHardwareAcceleration();
  app.commandLine.appendSwitch('disable-gpu');
  app.commandLine.appendSwitch('disable-gpu-compositing');
  app.commandLine.appendSwitch('disable-gpu-sandbox');
  app.commandLine.appendSwitch('disable-gpu-process-sandbox');
  app.commandLine.appendSwitch('disable-dev-shm-usage');
  app.commandLine.appendSwitch('no-sandbox');
  app.commandLine.appendSwitch('disable-unsafe-util');
  app.commandLine.appendSwitch('in-process-gpu');
  app.commandLine.appendSwitch('disable-direct-gl');
  app.commandLine.appendSwitch('use-gl=swiftshader');
  app.commandLine.appendSwitch('disable-software-rasterizer');
  app.commandLine.appendSwitch('disable-gpu-shader-disk-cache');
}

let mainWindow: BrowserWindow | null = null;

function createWindow(): void {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    minWidth: 1024,
    minHeight: 700,
    title: 'FunctionFly Studio',
    backgroundColor: '#0a0a0f',
    show: false,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: false,
    },
  });

  mainWindow.once('ready-to-show', () => {
    mainWindow?.show();
  });

  const studioUrl = isDev
    ? 'http://localhost:3000/studio'
    : `file://${path.join(__dirname, '../web/dashboard/dist/index.html')}`;

  if (isDev) {
    mainWindow.loadURL(studioUrl);
    mainWindow.webContents.openDevTools();
  } else {
    mainWindow.loadURL(studioUrl);
  }

  mainWindow.webContents.setWindowOpenHandler(({ url }: { url: string }) => {
    if (url.startsWith('http')) {
      shell.openExternal(url);
    }
    return { action: 'deny' };
  });

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

ipcMain.on('window-minimize', () => {
  mainWindow?.minimize();
});

ipcMain.on('window-maximize', () => {
  if (mainWindow?.isMaximized()) {
    mainWindow.unmaximize();
  } else {
    mainWindow?.maximize();
  }
});

ipcMain.on('window-close', () => {
  mainWindow?.close();
});

if (app) {
  app.whenReady().then(() => {
    createWindow();

    app.on('activate', () => {
      if (BrowserWindow.getAllWindows().length === 0) {
        createWindow();
      }
    });
  });

  app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
      app.quit();
    }
  });

  app.on('web-contents-created', (_event: Electron.Event, contents: Electron.WebContents) => {
    contents.on('will-navigate', (event: Electron.Event, navigationUrl: string) => {
      try {
        const parsedUrl = new URL(navigationUrl);
        const allowedOrigins = ['http://localhost:3000', 'file://'];
        const isAllowed = allowedOrigins.some((origin) => parsedUrl.origin === origin || navigationUrl.startsWith(origin));
        if (!isAllowed) {
          event.preventDefault();
        }
      } catch {
        event.preventDefault();
      }
    });
  });
}