import { Link, isAllowedUri } from '@tiptap/extension-link'
import { MarkdownSerializer, defaultMarkdownSerializer } from 'prosemirror-markdown'

import markdownItAttrs from 'markdown-it-attrs'
import type { Editor } from '@tiptap/vue-3'
import LinkChoser from './LinkChoser.vue'

// // https://snyk.io/advisor/npm-package/prosemirror-markdown/functions/prosemirror-markdown.MarkdownSerializer
// const serializer = new MarkdownSerializer(defaultMarkdownSerializer.nodes, {
//   ...defaultMarkdownSerializer.marks,
//   link: {
//     open(state, mark, parent, index) {
//       mark.attrs.title = 'Der Titel'
//       //   state.inAutolink = isPlainURL(mark, parent, index)
//       return '[' // state.inAutolink ? '<' : '['
//     },
//     close(state, mark, parent, index) {
//       //   const { inAutolink } = state
//       //   state.inAutolink = undefined
//       //   return inAutolink
//       //     ? '>'
//       //     : '](' + mark.attrs.href.replace(/[\(\)"]/g, '\\$&') + (mark.attrs.title ? ` "${mark.attrs.title.replace(/"/g, '\\"')}"` : '') + ')'
//       return (
//         ']('
//         + mark.attrs.href.replace(/[\(\)"]/g, '\\$&')
//         + (mark.attrs.title ? ` "${mark.attrs.title.replace(/"/g, '\\"')}"` : '')
//         + ')'
//         + (mark.attrs.target ? ` {target="${mark.attrs.target.replace(/"/g, '\\"')}"}` : '')
//       )
//     },
//     mixable: true,
//   },
//   italic: { open: '_', close: '_', mixable: true, expelEnclosingWhitespace: true },
// })

// // https://snyk.io/advisor/npm-package/prosemirror-markdown/functions/prosemirror-markdown.MarkdownSerializer
// const serializer = new MarkdownSerializer(defaultMarkdownSerializer.nodes, {
//   ...defaultMarkdownSerializer.marks,
//   bold: {
//     // creates a bold alias for the strong mark converter
//     ...defaultMarkdownSerializer.marks.strong,
//   },
//   link: {
//     open(state, mark, parent, index) {
//       mark.attrs.title = 'Der Titel'
//       //   state.inAutolink = isPlainURL(mark, parent, index)
//       return '[' // state.inAutolink ? '<' : '['
//     },
//     close(state, mark, parent, index) {
//       //   const { inAutolink } = state
//       //   state.inAutolink = undefined
//       //   return inAutolink
//       //     ? '>'
//       //     : '](' + mark.attrs.href.replace(/[\(\)"]/g, '\\$&') + (mark.attrs.title ? ` "${mark.attrs.title.replace(/"/g, '\\"')}"` : '') + ')'
//       return (
//         '](' +
//         mark.attrs.href.replace(/[\(\)"]/g, '\\$&') +
//         (mark.attrs.title ? ` "${mark.attrs.title.replace(/"/g, '\\"')}"` : '') +
//         ')' +
//         (mark.attrs.target ? ` {target="${mark.attrs.target.replace(/"/g, '\\"')}"}` : '')
//       )
//     },
//     mixable: true,
//   },
//   italic: { open: '_', close: '_', mixable: true, expelEnclosingWhitespace: true },
// })

export const CustomLink = Link.extend({

  addAttributes() {
    // console.log('link.addAttributes', this.parent?.())
    return {
      ...this.parent?.(),
      // href: {
      //   default: null,
      //   parseHTML(element) {
      //     return element.getAttribute('href')
      //   },
      // },
      // target: { default: this.options.HTMLAttributes.target },
      title: { default: this.options.HTMLAttributes.title },
      // class: { default: this.options.HTMLAttributes.class },
      // rel: { default: this.options.HTMLAttributes.rel },
      id: { default: this.options.HTMLAttributes.class },
    }
  },
  addStorage() {
    return {
      markdown: {
        serialize: {
          open(state, mark, parent, index) {
            // console.log('link.addStorage.markdown.serialize.open', mark)
            return '['
          },
          close(state, mark, parent, index) {
            let str = mark.attrs.id ? `#${mark.attrs.id}` : ''
            if (mark.attrs.class) {
              str += (str.length ? ' ' : '') + `.${mark.attrs.class}`
            }
            if (mark.attrs.target) {
              str += (str.length ? ' ' : '') + `target="${mark.attrs.target.replace(/"/g, '\\"')}"`
            }
            if (mark.attrs.rel) {
              str += (str.length ? ' ' : '') + `rel="${mark.attrs.rel.replace(/"/g, '\\"')}"`
            }
            // console.log('link close', mark, str)
            const result = (
              ']('
              + mark.attrs.href.replace(/[\(\)"]/g, '\\$&')
              + (mark.attrs.title ? ` "${mark.attrs.title.replace(/"/g, '\\"')}"` : '')
              + ')'
              + (str.length ? `{${str}}` : '')
            )
            // console.log('link.addStorage.markdown.serialize.close', result)
            return result
          },
          mixable: true,
        },
        parse: {
          setup(markdownit) {
            markdownit.use(markdownItAttrs, {
              // leave out to allow all attributes
              allowedAttributes: ['id', 'class', 'target', 'rel'],
            })
          },
        },
      },
    }
  },
})

export async function openLinkModal(editor: Editor) {
  const link = editor.getAttributes('link')
  // console.log('openLinkModal:', link)

  // const modal = useOverlay().create(LinkChoser)

  const data = await useOverlay().create(LinkChoser).open({
    href: link.href,
    rel: link.rel,
    id: link.id,
    class: link.class,
    target: link.target,
    title: link.title,
  })
  // const data = await instance

  // console.log('openLinkModal: data=', data)
  if (data === null) {
    editor.chain().focus().unsetLink().run()
  }
  if (typeof data === 'object') {
    console.log('link data =>', JSON.stringify(data))
    editor.chain().focus().extendMarkRange('link').toggleLink(data).run()
  }
// if (typeof data === 'object') {
//     console.log('data', data)
//     editor.value?.chain().focus().toggleLink({ href: data.href, target: data.target, rel: data.rel, class: data.class }).run()
//   }
// Expand selection to link mark and update attributes
// editor
//   .chain()
//   .extendMarkRange('link')
//   .updateAttributes('link', {
//     href: 'https://duckduckgo.com',
//   })
//   .run()
}
