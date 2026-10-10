<template>
  <!-- Sized against the round menu button at the other end of the strip: the
       two are what the bar is made of, and a name half the button's height read
       as a caption next to it. The cloud goes up with the letters — it is the
       last glyph of the word, not an ornament beside it.

       Which is why the cloud is 1em rather than a length: it was written as the
       same number of pixels as the type, and the two then had to be changed
       together and remembered about. In em it simply is the type's size, and
       the next change to the one is the change to the other. -->
  <h1 class="font-mono text-xl/6 font-black">
    ihle<span class="font-light opacity-70">ven</span><UIcon name="i-lucide-cloud" class="inline-block size-[1em] align-middle opacity-50" />
  </h1>
</template>
