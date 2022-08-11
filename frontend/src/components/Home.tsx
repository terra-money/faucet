import { formatValue } from '../libs/parse'
import { truncate } from '../libs/text'
import Button from './Button'
import styles from './Home.module.scss'
import { AstronautSVG, CloseSVG, FailedSVG, SuccessSVG } from './Svg'
import Tooltip from './Tooltip'
import { useWallet } from '@marsprotocol/wallet-connector'
import { CircularProgress } from '@material-ui/core'
import axios from 'axios'
import { useState } from 'react'
import { GoogleReCaptcha } from 'react-google-recaptcha-v3'
import { useTranslation } from 'react-i18next'

const Home = () => {
    const { t } = useTranslation()
    const { address = '', chainInfo } = useWallet()
    const [sending, setSending] = useState(false)
    const [error, setError] = useState(false)
    const [errorText, setErrorText] = useState('')
    const [sentAmount, setSentAmount] = useState(0)
    const [success, setSuccess] = useState(false)
    const [verified, setVerified] = useState(false)
    const [txUrl, setTxUrl] = useState('')
    const [captchaResponse, setCaptchaResponse] = useState('')

    const resetUI = () => {
        setError(false)
        setSuccess(false)
        setErrorText('')
        setSending(false)
    }

    const handleSubmit = () => {
        const faucetUrl = 'https://faucet.marsprotocol.io/claim'
        setSending(true)

        axios
            .post(faucetUrl, {
                address: address,
                denom: 'umars',
                response: captchaResponse,
            })
            .then((res) => {
                const { amount } = res.data
                setSentAmount(amount)
                const response =
                    res.data.response['tx_response'] || res.data.response

                if (response.code) {
                    setErrorText(
                        `Error: ${response.raw_log || `code: ${response.code}`}`
                    )
                    setSending(false)
                    setSuccess(false)
                    setError(true)
                } else {
                    setTxUrl(
                        `https://testnet-explorer.marsprotocol.io/transactions/${response.txhash}`
                    )
                    setSending(false)
                    setSuccess(true)
                    setError(false)
                }
            })
            .catch((err) => {
                let errText = err.message

                if (err.response) {
                    if (err.response.data) {
                        errText = err.response.data
                    } else {
                        switch (err.response.status) {
                            case 400:
                                errText = t('faucet.invalidRequest')
                                break
                            case 403:
                            case 429:
                                errText = t('faucet.tooManyRequests')
                                break
                            case 404:
                                errText = t('faucet.cannotConnect')
                                break
                            case 500:
                            case 502:
                            case 503:
                                errText = t('faucet.unavailable')
                                break
                            default:
                                errText = err.message
                        }
                    }
                }
                setErrorText(t('faucet.anErrorOccurred', { error: errText }))
                setSending(false)
                setError(true)
                setSuccess(false)
            })
    }

    return (
        <section className={styles.home}>
            <div className={styles.header}>
                {(success || error) && (
                    <div className={styles.close} onClick={resetUI}>
                        <CloseSVG />
                    </div>
                )}
                <h1 className={styles.title}>
                    {success
                        ? t('common.completed')
                        : error
                        ? t('common.errorEncountered')
                        : t('faucet.marsTestnetFaucet')}
                </h1>
                <div className={styles.tooltip}>
                    <Tooltip content={t('faucet.tooltip')} />
                </div>
            </div>
            {success ? (
                <div className={styles.content}>
                    <h2 className={styles.subtitle}>
                        {t('common.transactionSuccessful')}
                    </h2>
                    <div className={styles.status}>
                        <SuccessSVG />
                    </div>
                    <h3 className={styles.statusText}>
                        {t('faucet.tokensClaimed')}
                    </h3>
                    <dl className={styles.data}>
                        <dt className={styles.label}>{t('common.received')}</dt>
                        <dd className={styles.value}>
                            {formatValue(
                                sentAmount / 1000000,
                                2,
                                2,
                                true,
                                false,
                                ' MARS',
                                false,
                                false
                            )}
                        </dd>
                    </dl>
                    <p className={styles.copy}>
                        <a
                            href={txUrl}
                            target='_blank'
                            rel='noopener noreferrer'
                        >
                            {t('faucet.goToExplorer')}
                        </a>
                    </p>
                    <div className={styles.button}>
                        <Button
                            disabled={sending}
                            text={t('common.close')}
                            onClick={resetUI}
                        />
                    </div>
                </div>
            ) : error ? (
                <div className={styles.content}>
                    <h2 className={styles.subtitle}>
                        {t('common.transactionFailed')}
                    </h2>
                    <div className={styles.status}>
                        <FailedSVG />
                    </div>
                    <p className={styles.copy}>{errorText}</p>
                    <div className={styles.button}>
                        <Button
                            disabled={sending}
                            text={t('common.close')}
                            onClick={resetUI}
                        />
                    </div>
                </div>
            ) : (
                <div className={styles.content}>
                    <div className={styles.astronaut}>
                        <AstronautSVG />
                    </div>
                    <p className={styles.copy}>{t('faucet.welcomeMartian')}</p>
                    <p className={styles.copy}>{t('faucet.useTheFaucet')}</p>
                    <dl className={styles.data}>
                        <dt className={styles.label}>{t('common.chainId')}</dt>
                        <dd className={styles.value}>{chainInfo?.chainId}</dd>
                        <dt className={styles.label}>
                            {t('common.testnetAddress')}
                        </dt>
                        <dd className={styles.value}>
                            {truncate(address, [6, 4])}
                        </dd>
                    </dl>
                    <div className={styles.captcha}>
                        <GoogleReCaptcha
                            onVerify={(token: string) => {
                                setCaptchaResponse(token)
                                setVerified(true)
                            }}
                        />
                    </div>
                    <div className={styles.button}>
                        <Button
                            disabled={!verified || sending}
                            text={
                                sending ? (
                                    <CircularProgress />
                                ) : (
                                    t('faucet.sendMeTokens')
                                )
                            }
                            onClick={handleSubmit}
                        />
                    </div>
                </div>
            )}
        </section>
    )
}

export default Home
