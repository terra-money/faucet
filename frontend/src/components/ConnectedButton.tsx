import { formatValue, lookup } from '../libs/parse'
import { truncate } from '../libs/text'
import colors from '../styles/_assets.module.scss'
import Button from './Button'
import styles from './ConnectButton.module.scss'
import { CheckSVG, CopySVG, ExternalSVG, MarsSVG, WalletSVG } from './Svg'
import {
    ChainInfoID,
    useWallet,
    useWalletManager,
} from '@marsprotocol/wallet-connector'
import { CircularProgress, ClickAwayListener } from '@material-ui/core'
import { useCallback, useState } from 'react'
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
    const { name, address = '', chainInfo, walletBalances } = useWallet()
    const { t } = useTranslation()
    const [isCopied, setCopied] = useClipboard(address, {
        successDuration: 1000 * 5,
    })

    // ---------------
    // VARIABLES
    // ---------------
    const userBalances = walletBalances?.balances

    const userBalance =
        userBalances && userBalances.length ? userBalances[0].amount : '0'
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
                    {walletBalances ? (
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
                            color='inherit'
                            size={'0.9rem'}
                            className={styles.circularProgress}
                        />
                    )}
                </div>
            </button>
            {showDetails && (
                <ClickAwayListener onClickAway={onClickAway}>
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
                </ClickAwayListener>
            )}
        </>
    )
}

export default ConnectedButton
