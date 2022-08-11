interface ButtonStyleOverride {
    display?: string
    position?: 'relative' | 'absolute' | 'fixed'
    justifyContent?: string
    top?: string
    right?: string
    left?: string
    bottom?: string
    fontSize?: string
    marginLeft?: string
    marginRight?: string
    marginTop?: string
    marginBottom?: string
    padding?: string
    borderRadius?: string
    textTransform?: TextTransform
    backgroundImage?: string
    fontFamily?: string
    letterSpacing?: string
    backgroundColor?: string
    alignSelf?: string
    width?: string
    opacity?: string
    height?: string
}

type WrapperComponent = {
    children?: React.ReactNode
}
