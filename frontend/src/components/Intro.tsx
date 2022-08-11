import redBank from '../images/redbank.svg'
import ConnectButton from './ConnectButton'
import styles from './Intro.module.scss'
import { useTranslation } from 'react-i18next'

const Intro = () => {
    const { t } = useTranslation()

    return (
        <section className={styles.intro}>
            <div className={styles.content}>
                <img
                    src={redBank}
                    className={styles.redbank}
                    alt={t('faucet.marsTestnetFaucet')}
                />
                <h1 className={styles.title}>
                    {t('faucet.marsTestnetFaucet')}
                </h1>
                <h2 className={styles.subtitle}>
                    {t('faucet.welcomeMartian')}
                </h2>
                <p className={styles.copy}>
                    {t('faucet.connectYourKeplrWallet')}
                </p>
                <div className={styles.connect}>
                    <ConnectButton variant={styles.button} />
                </div>
            </div>
        </section>
    )
}

export default Intro
