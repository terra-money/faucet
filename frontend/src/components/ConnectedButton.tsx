import { formatValue, lookup } from '../libs/parse'
import { truncate } from '../libs/text'
import colors from '../styles/_assets.module.scss'
import Button from './Button'
import CircularProgress from './CircularProgress'
import styles from './ConnectButton.module.scss'
import { CheckSVG, CopySVG, ExternalSVG, MarsSVG, WalletSVG } from './Svg'
import {
    ChainInfoID,
    fetchBalances,
    useWallet,
    useWalletManager,
} from '@marsprotocol/wallet-connector'
import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import useClipboard from 'react-use-clipboard'

const ConnectedButton = () => {
    // ----------------
    // CONSTANTS
    // ----------------
    const FINDER_URL = 'https://www.mintscan.io/'

    // ---------------
    // EXTERNAL HOOKS
    // ---------------
    const { disconnect } = useWalletManager()
    const { name, address = '', chainInfo } = useWallet()
    const [userBalance, setUserBalance] = useState<string | undefined>()
    const { t } = useTranslation()
    const [isCopied, setCopied] = useClipboard(address, {
        successDuration: 1000 * 5,
    })

    // ---------------
    // VARIABLES
    // ---------------

    const [showDetails, setShowDetails] = useState(false)

    const viewOnFinder = useCallback(() => {
        window.open(
            `${FINDER_URL}${chainInfo?.chainName.toLocaleLowerCase()}/account/${address}`,
            '_blank'
        )
    }, [chainInfo, address])

    const onClickAway = useCallback(() => {
        setShowDetails(false)
    }, [])

    useEffect(() => {
        const interval = setInterval(async () => {
            const userBalances = await fetchBalances(
                address,
                chainInfo?.chainId
            )

            if (userBalances && userBalances.balances?.length) {
                setUserBalance(userBalances.balances[0].amount)
            }
        }, 3000)
        return () => clearInterval(interval)
    }, [address, chainInfo])

    return (
        <>
            {chainInfo?.chainId !== ChainInfoID.Mars1 && (
                <span className={styles.network}>{chainInfo?.chainId}</span>
            )}
            <button
                className={styles.button}
                onClick={() => {
                    setShowDetails(!showDetails)
                }}
            >
                <span className={styles.walletIcon}>
                    {chainInfo?.chainId === ChainInfoID.Mars1 ||
                    chainInfo?.chainId === ChainInfoID.MarsAres1 ? (
                        <MarsSVG className={styles.osmosisIcon} />
                    ) : (
                        <WalletSVG className={styles.walletIcon} />
                    )}
                </span>
                <span className={styles.address}>
                    {name ? name : truncate(address, [2, 4])}
                </span>
                <div className={styles.balance}>
                    {userBalance ? (
                        `${formatValue(
                            lookup(
                                Number(userBalance),
                                chainInfo?.stakeCurrency?.coinDenom || '',
                                chainInfo?.stakeCurrency?.coinDecimals || 6
                            ),
                            2,
                            2,
                            true,
                            false
                        )}`
                    ) : (
                        <CircularProgress
                            size={12}
                            className={styles.circularProgress}
                        />
                    )}
                </div>
            </button>
            {showDetails && (
                <>
                    <div className={styles.details}>
                        <div className={styles.detailsHeader}>
                            <div className={styles.detailsBalance}>
                                <p>
                                    <span className={styles.detailsDenom}>
                                        {chainInfo?.stakeCurrency?.coinDenom}
                                    </span>
                                    {formatValue(
                                        lookup(
                                            Number(userBalance),
                                            chainInfo?.stakeCurrency
                                                ?.coinDenom || '',
                                            chainInfo?.stakeCurrency
                                                ?.coinDecimals || 6
                                        ),
                                        2,
                                        2,
                                        true,
                                        false,
                                        false
                                    )}
                                </p>
                            </div>
                            <div className={styles.detailsButton}>
                                <Button
                                    text={t('common.disconnect')}
                                    color='secondary'
                                    onClick={disconnect}
                                />
                            </div>
                        </div>
                        <div className={styles.detailsBody}>
                            <p className={styles.addressLabel}>
                                {name ? `‘${name}’` : t('common.yourAddress')}
                            </p>
                            <p className={styles.address}>{address}</p>
                            <p className={styles.addressMobile}>
                                {truncate(address, [14, 14])}
                            </p>
                            <div className={styles.buttons}>
                                <button
                                    className={styles.copy}
                                    onClick={setCopied}
                                >
                                    <CopySVG color={colors.secondaryDark} />
                                    {isCopied ? (
                                        <>
                                            {t('common.copied')}{' '}
                                            <CheckSVG
                                                color={colors.secondaryDark}
                                            />
                                        </>
                                    ) : (
                                        <>{t('common.copy')}</>
                                    )}
                                </button>
                                <button
                                    className={styles.external}
                                    onClick={viewOnFinder}
                                >
                                    <ExternalSVG color={colors.secondaryDark} />{' '}
                                    {t('common.viewOnFinder')}
                                </button>
                            </div>
                        </div>
                    </div>
                    <div
                        className={styles.clickAway}
                        role='button'
                        onClick={onClickAway}
                    />
                </>
            )}
        </>
    )
}

export default ConnectedButton
