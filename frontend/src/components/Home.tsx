import { truncate } from '../libs/text'
import Button from './Button'
import styles from './Home.module.scss'
import { AstronautSVG } from './Svg'
import Tooltip from './Tooltip'
import { useWallet } from '@marsprotocol/wallet-connector'
import { useTranslation } from 'react-i18next'

const Home = () => {
    const { t } = useTranslation()
    const { address = '', chainInfo } = useWallet()
    return (
        <section className={styles.home}>
            <div className={styles.header}>
                <div className={styles.title}>
                    {t('faucet.marsTestnetFaucet')}
                </div>
                <div className={styles.tooltip}>
                    <Tooltip content={t('faucet.tooltip')} />
                </div>
            </div>
            <div className={styles.astronaut}>
                <AstronautSVG />
            </div>
            <p className={styles.copy}>{t('faucet.welcomeMartian')}</p>
            <p className={styles.copy}>{t('faucet.useTheFaucet')}</p>
            <dl className={styles.data}>
                <dt className={styles.label}>{t('common.chainId')}</dt>
                <dd className={styles.value}>{chainInfo?.chainId}</dd>
                <dt className={styles.label}>{t('common.testnetAddress')}</dt>
                <dd className={styles.value}>{truncate(address, [6, 4])}</dd>
            </dl>
            <div className={styles.button}>
                <Button text={t('faucet.sendMeTokens')} />
            </div>
        </section>
    )
}

export default Home
