import React from 'react'
import { render, RenderOptions } from '@testing-library/react'

interface WrapperProps {
  children: React.ReactNode
}

const AllProviders: React.FC<WrapperProps> = ({ children }) => {
  // Add providers here in the future (theme, router, etc.)
  return <>{children}</>
}

export function renderWithProviders(ui: React.ReactElement, options?: Omit<RenderOptions, 'wrapper'>) {
  return render(ui, { wrapper: AllProviders, ...options } as any)
}

export default renderWithProviders
