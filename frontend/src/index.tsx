import App from './App'
import CosmosWalletConnectProvider from './components/WalletConnectProvider'
import './styles/index.scss'
import { StrictMode, Suspense } from 'react'
import ReactDOM from 'react-dom/client'

ReactDOM.createRoot(document.getElementById('root')!).render(
    <Suspense fallback={null}>
        <StrictMode>
            <CosmosWalletConnectProvider>
                <App />
            </CosmosWalletConnectProvider>
        </StrictMode>
    </Suspense>
)
