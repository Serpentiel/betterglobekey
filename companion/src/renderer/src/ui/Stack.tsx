import type { ReactElement, CSSProperties, ElementType, ReactNode } from 'react'

interface StackProps {
  as?: ElementType
  direction?: 'row' | 'column'
  gap?: number
  align?: CSSProperties['alignItems']
  justify?: CSSProperties['justifyContent']
  wrap?: boolean
  flex?: CSSProperties['flex']
  className?: string
  style?: CSSProperties
  children: ReactNode
}

/**
 * Stack is the single layout primitive: a flexbox row or column with a gap. All
 * spacing between elements is expressed through it, so screens never hand-roll
 * fl/grid layout.
 */
export function Stack({
  as: Tag = 'div',
  direction = 'column',
  gap = 0,
  align,
  justify,
  wrap = false,
  flex,
  className,
  style,
  children,
}: StackProps): ReactElement {
  return (
    <Tag
      className={className}
      style={{
        display: 'flex',
        flexDirection: direction,
        gap: gap ? `${gap}px` : undefined,
        alignItems: align,
        justifyContent: justify,
        flexWrap: wrap ? 'wrap' : undefined,
        flex,
        ...style,
      }}
    >
      {children}
    </Tag>
  )
}
