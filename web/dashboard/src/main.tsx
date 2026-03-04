import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./i18n";
import { initSentry } from "./sentry";
import App from "./App.tsx";

initSentry();

// Load full stylesheets asynchronously after initial render
const loadFullStyles = () => {
  const link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = '/src/styles/index.css';
  document.head.appendChild(link);
};

// Load styles after initial render to avoid blocking LCP
setTimeout(loadFullStyles, 100);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
