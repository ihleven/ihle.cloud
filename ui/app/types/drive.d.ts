// HiDrive's own object model, as pkg/hi.Meta passes it through.
//
// Only the fields something renders are named; the provider sends more.

interface DriveMeta {
  name: string
  path: string
  /** "dir" or "file". */
  type?: string
  /** "directory", "image", "document", "code", … — what the browser branches on. */
  category?: string
  size?: number
  /** Unix seconds. Absent on entries the provider has no timestamp for. */
  mtime?: number
  ctime?: number
  nmembers?: number
  mime_type?: string
  members?: DriveMeta[]
  image?: { width: number, height: number, exif?: Record<string, string | number> }
  /** Only asked for on the entry itself; a member of a listing never carries them. */
  readable?: boolean
  writable?: boolean
}
