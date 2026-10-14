import type { TreeNode, ZipFile } from "./types"

export function buildFileTree(files: ZipFile[]): TreeNode[] {
  const root: TreeNode = { name: "", path: "", children: [] }

  const sorted = [...files].sort((a, b) => {
    if (a.is_directory !== b.is_directory) return a.is_directory ? -1 : 1
    return a.full_name.localeCompare(b.full_name)
  })

  for (const file of sorted) {
    const parts = file.full_name.split("/").filter(Boolean)
    let node = root
    let currentPath = ""
    for (let i = 0; i < parts.length; i++) {
      const name = parts[i]
      currentPath = currentPath ? `${currentPath}/${name}` : name
      const isLeaf = i === parts.length - 1
      let child = node.children.find((c) => c.name === name)
      if (!child) {
        child = {
          name,
          path: currentPath,
          file: isLeaf ? file : undefined,
          children: [],
        }
        node.children.push(child)
      }
      node = child
    }
  }

  return root.children
}

export interface FlatNode {
  key: string
  depth: number
  hasChildren: boolean
  file?: ZipFile
  name: string
  path: string
}

export function flattenTree(nodes: TreeNode[], depth = 0): FlatNode[] {
  const result: FlatNode[] = []
  for (const node of nodes) {
    result.push({
      key: node.file?.id ?? `dir:${node.path}`,
      depth,
      hasChildren: node.children.length > 0,
      file: node.file,
      name: node.name,
      path: node.path,
    })
    result.push(...flattenTree(node.children, depth + 1))
  }
  return result
}