export class ConfigLoader {
  // Naive YAML parser for simple key: value + nesting by indentation
  static parseYaml(yaml: string): any {
    const lines = yaml.split(/\r?\n/).filter((l) => l.trim().length > 0)
    const root: any = {}
    const stack: any[] = [root]
    const indentStack: number[] = [-1]

    for (const rawLine of lines) {
      const line = rawLine
      const indent = line.match(/^\s*/)?.[0].length ?? 0
      const content = line.trim()
      const colonIndex = content.indexOf(':')
      if (colonIndex === -1) continue
      const key = content.substring(0, colonIndex).trim()
      const value = content.substring(colonIndex + 1).trim()

      // Adjust current context according to indentation
      while (indent <= indentStack[indentStack.length - 1] && stack.length > 0) {
        stack.pop()
        indentStack.pop()
      }

      if (value === '' || value === undefined) {
        const obj: any = {}
        stack[stack.length - 1][key] = obj
        stack.push(obj)
        indentStack.push(indent)
      } else {
        let val: any = value
        if (/^-?\d+$/.test(value)) val = Number(value)
        else if (value === 'true') val = true
        else if (value === 'false') val = false
        stack[stack.length - 1][key] = val
      }
    }

    return root
  }
}
