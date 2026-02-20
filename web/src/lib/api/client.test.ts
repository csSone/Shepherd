import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ApiClient } from './client'

describe('ApiClient', () => {
  beforeEach(() => {
    // Ensure fetch is mocked for each test
    // @ts-ignore
    global.fetch = vi.fn()
  })

  it('GET fetch calls correct URL and returns parsed JSON', async () => {
    const fakeJson = { ok: true }
    // @ts-ignore
    (global.fetch as any).mockResolvedValue({ json: async () => fakeJson })
    const client = new ApiClient('https://api.example.com')
    const data = await client.get('/test')
    expect(data).toEqual(fakeJson)
    expect((global.fetch as any).mock.calls[0][0]).toBe('https://api.example.com/test')
    expect((global.fetch as any).mock.calls[0][1]).toMatchObject({ method: 'GET' })
  })

  it('POST fetch sends JSON body and returns parsed JSON', async () => {
    const fakeJson = { created: true }
    const mockFetch = vi.fn().mockResolvedValue({ json: async () => fakeJson })
    (global as any).fetch = mockFetch
    const client = new ApiClient('https://api.example.com')
    const body = { name: 'test' }
    const data = await client.post('/items', body)
    expect(data).toEqual(fakeJson)
    expect(mockFetch).toHaveBeenCalledWith(
      'https://api.example.com/items',
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
    )
  })
})
