import BigNumber from 'bignumber.js'
import numeral from 'numeral'

BigNumber.config({ EXPONENTIAL_AT: [-18, 20] })

type Formatter = (amount: string, symbol: string, decimals: number) => string

const rm = BigNumber.ROUND_HALF_CEIL

export const dp = (decimals: number, symbol?: string): number =>
    !symbol || symbol === 'uusd' ? 2 : decimals

export const lookup = (
    amount: number,
    symbol: string,
    decimals: number
): number => {
    const value = symbol
        ? new BigNumber(amount).div(10 ** decimals)
        : new BigNumber(amount)

    return value.dp(dp(decimals, symbol), rm).toNumber()
}

export const format: Formatter = (amount, symbol, decimals): string => {
    const value = new BigNumber(
        lookup(parseInt(amount || '0'), symbol, decimals)
    )
    const formatted = value.gte(10 ** decimals)
        ? numeral(value.div(1e4).integerValue(rm).times(1e4)).format(
              '0,0.[00]a'
          )
        : numeral(value).format('0,0.[000000]')

    return formatted.toUpperCase()
}

export const toAmount = (value: string, decimals: number): string =>
    value
        ? new BigNumber(value)
              .times(10 ** decimals)
              .integerValue()
              .toString()
        : '0'

export const magnify = (value: number, decimals: number): BigNumber | number =>
    value ? new BigNumber(value).times(10 ** decimals).integerValue() : 0

export const formatValue = (
    amount: number | string,
    minDecimals: number = 2,
    maxDecimals: number = 2,
    thousandSeparator: boolean = true,
    prefix: boolean | string = false,
    suffix: boolean | string = false,
    rounded: boolean = false,
    abbreviated: boolean = true
): string => {
    let numberOfZeroDecimals: number | null = null
    if (typeof amount === 'string') {
        const decimals = amount.split('.')[1] ?? null
        if (decimals && Number(decimals) === 0) {
            numberOfZeroDecimals = decimals.length
        }
    }
    let convertedAmount: number | string = +amount || 0

    const amountSuffix = abbreviated
        ? convertedAmount >= 1_000_000_000
            ? 'B'
            : convertedAmount >= 1_000_000
            ? 'M'
            : convertedAmount >= 1_000
            ? 'K'
            : false
        : ''

    const amountPrefix = prefix

    if (amountSuffix === 'B') {
        convertedAmount = Number(amount) / 1_000_000_000
    }
    if (amountSuffix === 'M') {
        convertedAmount = Number(amount) / 1_000_000
    }
    if (amountSuffix === 'K') {
        convertedAmount = Number(amount) / 1_000
    }

    if (rounded) {
        convertedAmount = convertedAmount.toFixed(maxDecimals)
    } else {
        const amountFractions = String(convertedAmount).split('.')
        if (maxDecimals > 0) {
            if (typeof amountFractions[1] !== 'undefined') {
                if (amountFractions[1].length >= maxDecimals) {
                    convertedAmount = `${
                        amountFractions[0]
                    }.${amountFractions[1].substr(0, maxDecimals)}`
                }
                if (amountFractions[1].length < minDecimals) {
                    convertedAmount = `${
                        amountFractions[0]
                    }.${amountFractions[1].padEnd(minDecimals, '0')}`
                }
            }
        } else {
            convertedAmount = amountFractions[0]
        }
    }

    if (thousandSeparator) {
        convertedAmount = Number(convertedAmount).toLocaleString('en', {
            useGrouping: true,
            minimumFractionDigits: minDecimals,
            maximumFractionDigits: maxDecimals,
        })
    }

    let returnValue = ''
    if (amountPrefix) {
        returnValue += amountPrefix
    }

    returnValue += convertedAmount

    // Used to allow for numbers like 1.0 or 3.00 (otherwise impossible with string to number conversion)
    if (numberOfZeroDecimals) {
        if (numberOfZeroDecimals < maxDecimals) {
            returnValue = Number(returnValue).toFixed(numberOfZeroDecimals)
        } else {
            returnValue = Number(returnValue).toFixed(maxDecimals)
        }
    }

    if (amountSuffix) {
        returnValue += amountSuffix
    }

    if (suffix) {
        returnValue += suffix
    }

    return returnValue
}
