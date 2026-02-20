import { describe, it, expect } from 'vitest'
import { ConfigLoader } from './configLoader'

describe('ConfigLoader', () => {
  it('parses simple YAML', () => {
    const yaml = 'server:\n  port: 3000'
    const result = ConfigLoader.parseYaml(yaml)
    expect(result.server.port).toBe(3000)
  })

  it('parses nested YAML with multiple levels', () => {
    const yaml = 'database:\n  host: localhost\n  port: 5432'
    const result = ConfigLoader.parseYaml(yaml)
    expect(result.database.host).toBe('localhost')
    expect(result.database.port).toBe(5432)
  })
})
