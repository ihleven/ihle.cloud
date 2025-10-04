import { Extension } from '@tiptap/core'
import { Document } from '@tiptap/extension-document'
import { Paragraph } from '@tiptap/extension-paragraph'
import type { ParagraphOptions } from '@tiptap/extension-paragraph'
import { Text } from '@tiptap/extension-text'
import type { HardBreakOptions } from '@tiptap/extension-hard-break'
import { HardBreak } from '@tiptap/extension-hard-break'
import type { HeadingOptions } from '@tiptap/extension-heading'
import { Heading } from '@tiptap/extension-heading'
import type { HighlightOptions } from '@tiptap/extension-highlight'
import type { HorizontalRuleOptions } from '@tiptap/extension-horizontal-rule'
import { HorizontalRule } from '@tiptap/extension-horizontal-rule'
import type { BulletListOptions } from '@tiptap/extension-bullet-list'
import { BulletList } from '@tiptap/extension-bullet-list'
import type { OrderedListOptions } from '@tiptap/extension-ordered-list'
import { OrderedList } from '@tiptap/extension-ordered-list'
import type { ListItemOptions } from '@tiptap/extension-list-item'
import { ListItem } from '@tiptap/extension-list-item'
import type { BlockquoteOptions } from '@tiptap/extension-blockquote'
import { Blockquote } from '@tiptap/extension-blockquote'
import type { CodeBlockOptions } from '@tiptap/extension-code-block'
import { CodeBlock } from '@tiptap/extension-code-block'
import type { ImageOptions } from '@tiptap/extension-image'
import { Image } from '@tiptap/extension-image'

import type { BoldOptions } from '@tiptap/extension-bold'
import { Bold } from '@tiptap/extension-bold'
import type { CodeOptions } from '@tiptap/extension-code'
import { Code } from '@tiptap/extension-code'
import { Highlight } from '@tiptap/extension-highlight'

import HighlightExt from './highlight'
import type { ItalicOptions } from '@tiptap/extension-italic'
import { Italic } from '@tiptap/extension-italic'
import type { StrikeOptions } from '@tiptap/extension-strike'
import { Strike } from '@tiptap/extension-strike'
import type { SubscriptExtensionOptions } from '@tiptap/extension-subscript'
import { Subscript } from '@tiptap/extension-subscript'
import type { SuperscriptExtensionOptions } from '@tiptap/extension-superscript'
import { Superscript } from '@tiptap/extension-superscript'
import type { TextStyleOptions } from '@tiptap/extension-text-style'
import TextStyle from '@tiptap/extension-text-style'
import type { UnderlineOptions } from '@tiptap/extension-underline'
import { Underline } from '@tiptap/extension-underline'

import type { HistoryOptions } from '@tiptap/extension-history'
import { History } from '@tiptap/extension-history'
import type { DropcursorOptions } from '@tiptap/extension-dropcursor'
import { Dropcursor } from '@tiptap/extension-dropcursor'
import { Gapcursor } from '@tiptap/extension-gapcursor'
import type { TypographyOptions } from '@tiptap/extension-typography'
import { Typography } from '@tiptap/extension-typography'
import { TextAlign } from '@tiptap/extension-text-align'

import type { LinkOptions } from '@tiptap/extension-link'
import { Link, openLinkModal } from './link'

export { openLinkModal }

export interface KitOptions {

  // Nodes
  document: false
  paragraph: Partial<ParagraphOptions>
  text: false
  hardBreak: Partial<HardBreakOptions>
  heading: Partial<HeadingOptions> | false
  horizontalRule: Partial<HorizontalRuleOptions> | false
  bulletList: Partial<BulletListOptions> | false
  orderedList: Partial<OrderedListOptions> | false
  listItem: Partial<ListItemOptions> | false
  blockquote: Partial<BlockquoteOptions> | false
  codeBlock: Partial<CodeBlockOptions> | false
  image: Partial<ImageOptions> | false

  // Marks
  bold: Partial<BoldOptions> | false
  code: Partial<CodeOptions> | false
  italic: Partial<ItalicOptions> | false
  highlight: Partial<HighlightOptions> | false
  strike: Partial<StrikeOptions> | false
  subscript: Partial<SubscriptExtensionOptions> | false
  superscript: Partial<SuperscriptExtensionOptions> | false
  textstyle: Partial<TextStyleOptions> | false
  underline: Partial<UnderlineOptions> | false

  link: Partial<LinkOptions> | false

  // functionality
  dropcursor: Partial<DropcursorOptions> | false
  gapcursor: false
  history: Partial<HistoryOptions> | false
  typography: Partial<TypographyOptions> | false

}

export const Kit = Extension.create<KitOptions>({
  name: 'kit',

  addExtensions() {
    const extensions = [
      Document,
      Paragraph.configure(this.options.paragraph),
      Text,
      HardBreak.configure(this.options.hardBreak),
      this.options.heading ? Heading.configure(this.options.heading) : Heading.configure(),
      TextAlign.configure({
        types: ['heading', 'paragraph'],
      }),
    ]

    if (this.options.horizontalRule !== false) {
      extensions.push(HorizontalRule.configure(this.options.horizontalRule))
    }
    if (this.options.bulletList !== false) {
      extensions.push(BulletList.configure(this.options.bulletList))
    }
    if (this.options.orderedList !== false) {
      extensions.push(OrderedList.configure(this.options.orderedList))
    }
    if (this.options.listItem !== false) {
      extensions.push(ListItem.configure(this.options.listItem))
    }
    if (this.options.blockquote !== false) {
      extensions.push(Blockquote.configure(this.options.blockquote))
    }
    if (this.options.codeBlock !== false) {
      extensions.push(CodeBlock.configure(this.options.codeBlock))
    }
    if (this.options.image !== false) {
      extensions.push(Image.configure(this.options.image))
    }

    //////////////////

    if (this.options.bold !== false) {
      extensions.push(Bold.configure(this.options.bold))
    }
    if (this.options.code !== false) {
      extensions.push(Code.configure(this.options.code))
    }
    if (this.options.highlight !== false) {
      extensions.push(this.options.highlight.old ? HighlightExt.configure(this.options.highlight) : Highlight.configure(this.options.highlight))
    }
    if (this.options.italic !== false) {
      extensions.push(Italic.configure(this.options.italic))
    }
    if (this.options.strike !== false) {
      extensions.push(Strike.configure(this.options.strike))
    }
    if (this.options.subscript !== false) {
      extensions.push(Subscript.configure(this.options.subscript))
    }
    if (this.options.superscript !== false) {
      extensions.push(Superscript.configure(this.options.superscript))
    }
    if (this.options.textstyle !== false) {
      extensions.push(TextStyle.configure(this.options.textstyle))
    }
    if (this.options.underline !== false) {
      extensions.push(Underline.configure(this.options.underline))
    }
    if (this.options.link !== false) {
      // extensions.push(Link.configure(this.options.link))
      extensions.push(Link.configure({
        openOnClick: false,
        HTMLAttributes: { rel: null, target: null },
      }))
    }

    //

    if (this.options.dropcursor !== false) {
      extensions.push(Dropcursor.configure(this.options.dropcursor))
    }

    if (this.options.gapcursor !== false) {
      extensions.push(Gapcursor.configure(this.options.gapcursor))
    }

    if (this.options.history !== false) {
      extensions.push(History.configure(this.options.history))
    }

    if (this.options.typography !== false) {
      extensions.push(Typography.configure(this.options.typography))
    }

    return extensions
  },
})
