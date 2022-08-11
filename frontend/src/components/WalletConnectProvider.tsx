import KeplrImage from '../images/wallets/keplr-wallet-extension.png'
import WalletConnectImage from '../images/wallets/keplr-walletconnect.png'
import buttonStyles from './Button.module.scss'
import styles from './WalletConnectProvider.module.scss'
import {
    ChainInfoID,
    WalletManagerProvider,
    WalletType,
} from '@marsprotocol/wallet-connector'
import { CircularProgress } from '@material-ui/core'
import i18next from 'i18next'
import { FC, useState } from 'react'
import { Trans } from 'react-i18next'

const CosmosWalletConnectProvider: FC<WrapperComponent> = ({ children }) => {
    const [i18nInit, setI18nInit] = useState(false)

    i18next.on('initialized', () => {
        setI18nInit(true)
    })

    if (i18nInit) {
        return (
            <WalletManagerProvider
                defaultChainId={ChainInfoID.MarsAres1}
                enablingStringOverride={
                    <Trans i18nKey='global.connectingToWallet' />
                }
                walletConnectClientMeta={{
                    name: 'Mars - Tesnet Faucet',
                    description: 'Mars Hub Testnet Faucet.',
                    url: 'https://faucet.marsprotocol.io',
                    icons: ['https://marsprotocol.io/favicon.svg'],
                }}
                classNames={{
                    modalContent: styles.content,
                    modalOverlay: styles.overlay,
                    modalHeader: styles.header,
                    modalCloseButton: styles.close,
                    walletList: styles.list,
                    walletImage: styles.image,
                    walletInfo: styles.info,
                    walletName: styles.name,
                    walletDescription: styles.description,
                    textContent: styles.text,
                }}
                enabledWalletTypes={[
                    WalletType.Keplr,
                    WalletType.WalletConnectKeplr,
                ]}
                walletMetaOverride={{
                    [WalletType.Keplr]: {
                        description: (
                            <Trans i18nKey='global.keplrBrowserExtension' />
                        ),
                        imageUrl: KeplrImage,
                    },
                    [WalletType.WalletConnectKeplr]: {
                        name: <Trans i18nKey='global.walletConnect' />,
                        description: (
                            <Trans i18nKey='global.walletConnectDescription' />
                        ),
                        imageUrl: WalletConnectImage,
                    },
                }}
                localStorageKey='walletConnection'
                renderLoader={() => (
                    <div className={styles.loader}>
                        <CircularProgress size={20} />
                    </div>
                )}
                closeIcon={<div />}
                enablingMeta={{
                    text: <Trans i18nKey='global.walletResetText' />,
                    textClassName: styles.text,
                    buttonText: <Trans i18nKey='global.walletReset' />,
                    buttonClassName: ` ${buttonStyles.button} ${buttonStyles.primary} ${buttonStyles.small} ${buttonStyles.solid} ${styles.button}`,
                    contentClassName: styles.enableContent,
                }}
            >
                {children}
            </WalletManagerProvider>
        )
    } else {
        return null
    }
}

export default CosmosWalletConnectProvider
