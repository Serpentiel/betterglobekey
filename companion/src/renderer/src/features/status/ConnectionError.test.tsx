// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { ConnectionError } from './ConnectionError'

afterEach(cleanup)

it('shows the failure message and retries on click', () => {
  const onRetry = vi.fn()
  render(<ConnectionError message="control socket missing" onRetry={onRetry} />)

  expect(screen.getByText(/control socket missing/)).toBeTruthy()

  fireEvent.click(screen.getByRole('button', { name: /Retry/ }))
  expect(onRetry).toHaveBeenCalledTimes(1)
})
