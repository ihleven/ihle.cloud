<template>
  <div>
    <bubble-menu
      v-if="editor"

      :editor="editor"
      :tippy-options="{ duration: 100 }"
      class="flex items-center"
    >
      <button
        v-for="m in bubble"
        :key="m"
        class="flex items-center bg-inverted p-1 py-0.5 text-inverted ring-2 ring-inverted first:rounded-s-full last:rounded-e-full hover:bg-default hover:text-default"
        :class="{ 'bg-muted font-black text-muted': isActive(m) }"
        @click="execmd(m)"
      >
        <UIcon
          v-if="icon(m)"
          :name="icon(m)"
          class="size-5 border border-black"
        />
        <template v-else>
          {{ m }}
        </template>
      </button>
    </bubble-menu>

    <floating-menu
      v-if="editor"
      :editor="editor"
      :tippy-options="{ duration: 100 }"
    >
      <button
        v-for="m in floating"
        :key="m"
        class="mr-4"
        :class="{ 'is-active': isActive(m) }"
        @click="execmd(m)"
      >
        <Icon
          v-if="icon(m)"
          :name="icon(m)"
          class="h-5 w-5"
        />
        <template v-else>
          {{ m }}
        </template>
      </button>
    </floating-menu>

    <editor-content
      :editor="editor"
      class="overflow-y-auto"
    />
  </div>
</template>

<script lang="ts" setup>
import type { Editor } from '@tiptap/vue-3'

import { useEditor, EditorContent, BubbleMenu, FloatingMenu } from '@tiptap/vue-3'

import { Markdown } from 'tiptap-markdown'
import { Kit, openLinkModal } from '~/tiptap'
// import { ImagekitDialog } from '#components'
// import { Figure } from '@/tiptap/experiments/figure'
// import { TravelguideImage } from '@/tiptap/travelguide/image/TravelguideImage'
// import { TravelguideToc } from '@/tiptap/travelguide/toc/TravelguideToc'
// import { TravelguideRecom } from '@/tiptap/travelguide/recommender/TravelguideRecom'
// import { TravelguideAccom } from '@/tiptap/travelguide/accommodation/TravelguideAccom'
// import { TravelguideHighlightBox } from '@/tiptap/travelguide/highlightbox/TravelguideHighlightBox'
// import { TravelguideIframe } from '@/tiptap/travelguide/iframe/TravelguideIframe'
// import { getHierarchicalIndexes, TableOfContents } from '@tiptap/extension-table-of-contents'

const props = defineProps<{
  modelValue: string
  bubbleMenu: string
  floatingMenu: string
  reactive?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [string]
}>()

const bubble = computed(() => (props.bubbleMenu ? props.bubbleMenu.split(',') : []))
const floating = computed(() => (props.floatingMenu ? props.floatingMenu.split(',') : []))

const model = computed({
  get: () => {
    return props.modelValue
  },
  set: (val: string) => {
    emit('update:modelValue', val)
  },
})

const editor = useEditor({
  // enableContentCheck: true,
  content: model.value,
  // onContentError({ editor, error, disableCollaboration }) {
  //   console.log('tiptap error: ', editor, error, disableCollaboration)
  // },
  extensions: [
    Kit.configure({
      highlight: { multicolor: true }, // HTMLAttributes: { class: 'my-custom-highlight-class' } },
      // underline: false,
      link: {
        // HTMLAttributes: { class: 'my-custom-class', rel: null, target: '_blank' },
        autolink: true,
        openOnClick: false,
      },
      typography: false,
    }),
    Markdown,
    // Figure,
    // TravelguideImage,
    // TravelguideToc,
    // TravelguideRecom,
    // TravelguideHighlightBox,
    // TravelguideAccom,
    // TravelguideIframe,
    // TableOfContents.configure({
    //   getIndex: getHierarchicalIndexes,
    //   // onUpdate: (content) => {
    //   //   console.log('toc:', content)
    //   // },
    // }),
  ],
  editorProps: {
    attributes: { class: 'focus:outline-none' },
  },
  onUpdate: () => {
    model.value = editor.value?.storage.markdown.getMarkdown()
  },
})

watch(() => props.modelValue, (nv) => {
  const ov = editor.value?.storage.markdown.getMarkdown()
  if (props.reactive && nv !== ov) {
    editor.value?.commands.setContent(nv)
    const nvstr = nv.length > 20 ? nv.slice(0, 20) : nv
    const ovstr = ov.length > 20 ? ov.slice(0, 20) : ov
    console.log(`tiptap watcher => ${nvstr}... !== ${ovstr}...`)
  }
})

onBeforeUnmount(() => {
  editor.value?.destroy()
})

// state fur isActive, label unused
const commands: Record<string, { label: string, state: string, command: string, icon: string, options?: object }> = {
  B: { icon: 'ci:bold', label: 'B', state: 'bold', command: 'toggleBold' },
  C: { icon: 'ci:code', label: 'C', state: 'code', command: 'toggleCode' },
  I: { icon: 'ci:italic', label: 'I', state: 'italic', command: 'toggleItalic' },
  S: { icon: 'ci:strikethrough', label: 'S', state: 'strike', command: 'toggleStrike' },
  U: { icon: 'ci:underline', label: 'U', state: 'underline', command: 'toggleUnderline' },
  H: { icon: 'i-lucide-highlighter', label: 'hilight', state: 'hilight', command: 'toggleHighlight' },
  sub: { icon: 'i-lucide-subscript', label: 'sub', state: 'subscript', command: 'toggleSubscript' },
  sup: { icon: 'i-lucide-superscript', label: 'sup', state: 'superscript', command: 'toggleSuperscript' },
  // style: { icon: 'ci:style', label: 'style', state: 'superscript', command: 'toggleSuperscript' },
  link: { icon: 'ci:link', label: 'link', state: 'link', func: openLinkModal },

  h1: { icon: 'ci:heading-h1', label: 'h1', state: 'heading', command: 'toggleHeading', options: { level: 1 } },
  h2: { icon: 'ci:heading-h2', label: 'h2', state: 'heading', command: 'toggleHeading', options: { level: 2 } },
  h3: { icon: 'ci:heading-h3', label: 'h3', state: 'heading', command: 'toggleHeading', options: { level: 3 } },
  h4: { icon: 'ci:heading-h4', label: 'h4', state: 'heading', command: 'toggleHeading', options: { level: 4 } },
  h5: { icon: 'ci:heading-h5', label: 'h5', state: 'heading', command: 'toggleHeading', options: { level: 5 } },
  h6: { icon: 'ci:heading-h6', label: 'h6', state: 'heading', command: 'toggleHeading', options: { level: 6 } },
  ul: { icon: 'ci:list-unordered', label: 'ul', state: 'bulletList', command: 'toggleBulletList' },
  ol: { icon: 'ci:list-ordered', label: 'ol', state: 'orderedList', command: 'toggleOrderedList' },
  img: { icon: 'ci:image', label: 'img', state: 'image', func: imageSelectorModal },
  alignLeft: { icon: 'i-lucide-align-left', state: 'textAlign', command: 'setTextAlign', options: 'left' },
  alignRight: { icon: 'i-lucide-align-right', state: 'textAlign', command: 'setTextAlign', options: 'right' },
  alignCenter: { icon: 'i-lucide-align-center', state: 'textAlign', command: 'setTextAlign', options: 'center' },
  alignJustify: { icon: 'i-lucide-align-justify', state: 'textAlign', command: 'setTextAlign', options: 'justify' },
  undo: { label: 'ci:undo', state: 'undo', command: 'undo' },
  redo: { label: 'ci:redo', state: 'redo', command: 'redo' },

  // figure: { func: excFuncFigure },
  // // html figure aus @/tiptap/experiments/figure
  // setFigure: { label: 'figure', state: 'figure', command: 'setFigure', options: { caption: 'caption', src: 'https://www.interhome.at/upload/travelguide/oesterreich-wandern-wachau.jpg' } },
  // imageToFigure: { label: 'imageToFigure', state: 'imageToFigure', command: 'imageToFigure' },
  // figureToImage: { label: 'figureToImage', state: 'figureToImage', command: 'figureToImage' },

  // // travelguide
  // captionedImage: { command: 'setCaptionedImage' },
  // toc: { command: 'setToc' },
  // recommender: { command: 'setRecom' },
  // iframe: { command: 'setIframe' },
  // accommodation: { command: 'setAccom', options: { code: 'AT9981.649.3' } },
  // highlightBox: { command: 'setHighlightBox' },
}

function execmd(cmd: string) {
  const command = commands[cmd] ? commands[cmd] : {}
  console.log('execmd', cmd, command)
  command.func ? command.func(editor.value) : editor.value.chain().focus()[command.command](command.options).run()
}

function isActive(cmd: string): boolean {
  const command = commands[cmd] ? commands[cmd] : {}
  return editor.value ? editor.value.isActive(command.state, command.options) : false
}

function icon(cmd: string): string {
  const command = commands[cmd] ? commands[cmd] : {}
  return command.icon ? command.icon : ''
}

// cmd img
async function imageSelectorModal(editor: Editor) {
  const { src, alt, title } = editor.getAttributes('image')
  console.log('imageSelectorModal:', src, alt, title)

  const modal = useOverlay().create(ImagekitDialog, { props: { src, alt, title } })

  const data: { src: string, alt: string, title: string } | boolean = await modal.open()
  console.log('imageSelectorModal: src=', data.src)

  if (typeof data === 'object') {
    editor.chain().focus().setImage(data).run()
  }
}

// cmd figure => in extension
async function excFuncFigure(editor: Editor) {
  const modal = useOverlay().create(ImagekitDialog, { props: { } })

  const instance = await modal.open({})

  const data = await instance.result
  console.log('figure: data=', data)
  if (typeof data === 'object') {
    editor.chain().focus().setFigure({ src: data.src, alt: data.alt, title: data.title, caption: 'the caption' }).run()
  }
}

defineExpose({
  editor,
  execmd,
  icon,

})
</script>

<style scoped lang="css">
.ProseMirror-gapcursor {
  display: none;
  pointer-events: none;
  position: absolute;
}

.ProseMirror-gapcursor:after {
  content: "";
  display: block;
  position: absolute;
  top: -2px;
  width: 20px;
  border-top: 1px solid black;
  animation: ProseMirror-cursor-blink 1.1s steps(2, start) infinite;
}

@keyframes ProseMirror-cursor-blink {
  to {
    visibility: hidden;
  }
}

.ProseMirror-focused .ProseMirror-gapcursor {
  display: block;
}
</style>
