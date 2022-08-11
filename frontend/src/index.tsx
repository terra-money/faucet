import App from './App'
import CosmosWalletConnectProvider from './components/WalletConnectProvider'
import './styles/index.scss'
import { StrictMode, Suspense } from 'react'
import ReactDOM from 'react-dom/client'
import { GoogleReCaptchaProvider } from 'react-google-recaptcha-v3'

ReactDOM.createRoot(document.getElementById('root')!).render(
    <Suspense fallback={null}>
        <StrictMode>
            <CosmosWalletConnectProvider>
                <GoogleReCaptchaProvider reCaptchaKey='6LdtqBQTAAAAAI-G1Zg2GqnYEoMWKAeq_GftuQI2'>
                    <App />
                </GoogleReCaptchaProvider>
            </CosmosWalletConnectProvider>
        </StrictMode>
    </Suspense>
)
