import styles from './CircularProgress.module.scss'

interface Props {
    color?: string
    size?: number
    className?: string
}

const CircularProgress = ({
    color = '#FFFFFF',
    size = 20,
    className,
}: Props) => {
    return (
        <div
            style={{ width: `${size}px`, height: `${size}px` }}
            className={
                className ? `${className} ${styles.loader}` : styles.loader
            }
        >
            <div
                style={{
                    borderWidth: `${size / 10}px`,
                    borderColor: `${color} transparent transparent transparent`,
                }}
                className={styles.element}
            />
            <div
                style={{
                    borderWidth: `${size / 10}px`,
                    borderColor: `${color} transparent transparent transparent`,
                }}
                className={styles.element}
            />
            <div
                style={{
                    borderWidth: `${size / 10}px`,
                    borderColor: `${color} transparent transparent transparent`,
                }}
                className={styles.element}
            />
            <div
                style={{
                    borderWidth: `${size / 10}px`,
                    borderColor: `${color} transparent transparent transparent`,
                }}
                className={styles.element}
            />
        </div>
    )
}

export default CircularProgress
