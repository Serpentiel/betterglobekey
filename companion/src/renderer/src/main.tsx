import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import { App } from './App'
import './styles/theme.css'

// Prevent dropping a file onto the window from navigating away from the app.
for (const event of ['dragover', 'drop']) {
  window.addEventListener(event, (e) => e.preventDefault())
}

const container = document.getElementById('root')

if (!container) {
  throw new Error('root element not found')
}

createRoot(container).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
