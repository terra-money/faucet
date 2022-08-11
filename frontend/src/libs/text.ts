export function truncate(
    text: string = '',
    [h, t]: [number, number] = [6, 6]
): string {
    const head = text.slice(0, h)
    if (t === 0) return text.length > h + t ? head + '...' : text
    const tail = text.slice(-1 * t, text.length)
    if (h === 0) return text.length > h + t ? '...' + tail : text
    return text.length > h + t ? [head, tail].join('...') : text
}
