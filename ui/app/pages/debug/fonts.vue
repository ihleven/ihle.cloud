<template>
  <article class="min-h-full w-full bg-elevated">
    <header class="flex items-baseline gap-2 border-b border-accented px-3 py-2">
      <h1 class="text-sm font-semibold text-highlighted">Schriften</h1>
      <span class="text-xs text-muted">iA Writer, drei Familien</span>
    </header>

    <div class="mx-auto max-w-3xl space-y-8 px-4 py-6">
      <section class="space-y-3 text-sm text-default">
        <p>
          Beide Schriften werden von hier ausgeliefert, nicht zur Laufzeit von einem Anbieter geholt — sie kommen nur
          auf verschiedenen Wegen ins Haus. <strong>Inter</strong> steckt im Paket
          <code class="font-mono text-xs">@fontsource-variable/inter</code>: eine Abhängigkeit wie jede andere, mit
          einer Fassung im Lockfile und einem Weg zur nächsten. <strong>iA Writer</strong> liegt unter
          <code class="font-mono text-xs">app/assets/fonts/ia/</code>, weil es das Paket nur mit festen Schnitten gibt
          und die variablen Achsen mehr wert sind als das Update. Unter <code class="font-mono text-xs">assets/</code>
          und nicht <code class="font-mono text-xs">public/</code>, damit der Build die Dateien mit einer Prüfsumme im
          Namen ausliefert: was so heißt, darf für immer zwischengespeichert werden, weil ein Austausch den Namen
          ändert. Beide stehen unter der SIL Open Font License; die Lizenz
          liegt jeweils bei den Dateien, denn sie gehört zur Schrift und nicht zum Verzeichnis.
        </p>
        <p>
          Beim Bauen wird nichts aus dem Netz geladen. Das wäre die dritte Möglichkeit — ein Anbieter liefert Inter
          beim Build —, kostet aber einen Build, der ans Netz muss, und eine Fassung, die niemand festgelegt hat.
        </p>
        <p>
          Es sind <strong>variable</strong> Schnitte: eine Datei je Familie und Lage deckt einen Bereich von
          Strichstärken ab. Deshalb sind es sechs Dateien statt vierundzwanzig — und deshalb lässt sich unten jede
          Zwischenstärke anzeigen, nicht nur Regular und Bold.
        </p>
      </section>

      <!-- The specimen, and the markup that makes it, from one source: the
           fragment below is rendered and printed, so the page cannot show one
           thing and claim another. -->
      <section class="space-y-2">
        <h2 class="text-sm font-semibold text-highlighted">Die drei Familien</h2>
        <p class="text-sm text-default">
          Sie unterscheiden sich darin, wie viel Platz ein Zeichen bekommt. <strong>Mono</strong> gibt jedem
          denselben. <strong>Duo</strong> gibt den wirklich breiten Buchstaben — <code
            class="font-mono text-xs"
          >m</code>, <code class="font-mono text-xs">w</code> — einen doppelten, damit sie nicht gequetscht wirken.
          <strong>Quattro</strong> kennt vier Breiten und liest sich dadurch fast wie eine Proportionalschrift; sie
          ist für Fließtext gedacht.
        </p>
        <p class="text-sm text-default">
          Am Wort sieht man das nicht, an den Wiederholungen schon: in Mono sind alle vier Blöcke gleich lang, in Duo
          werden <code class="font-mono text-xs">mmmm</code> und <code class="font-mono text-xs">wwww</code> länger,
          in Quattro sind alle vier verschieden.
        </p>

        <!-- eslint-disable-next-line vue/no-v-html -- the fragment is a constant in this file -->
        <div class="rounded-sm border border-accented bg-default p-4 text-lg/loose" v-html="fragment" />
      </section>

      <section class="space-y-2">
        <h2 class="text-sm font-semibold text-highlighted">Warum „Handgloves“ und „0123“</h2>
        <p class="text-sm text-default">
          <strong>Handgloves</strong> ist ein Musterwort aus dem Schriftsatz: es enthält Oberlängen
          (<code class="font-mono text-xs">H d l</code>), eine Unterlänge (<code class="font-mono text-xs">g</code>),
          Rundungen (<code class="font-mono text-xs">o e a</code>) und eine Schräge
          (<code class="font-mono text-xs">v</code>) — an einem Wort sieht man fast alles, was eine Schrift ausmacht.
          Der Klassiker ist „Hamburgefonstiv“; „Handgloves“ ist die kürzere Variante.
        </p>
        <p class="text-sm text-default">
          <strong>0123</strong> zeigt die Ziffern: ob sie auf einer Linie stehen, ob die Null geschnitten ist (hier
          ja) und — bei einer dicktengleichen Schrift — ob Ziffern gleich breit sind, was Zahlenkolonnen untereinander
          hält.
        </p>
        <p class="text-sm text-default">
          Was hier <em>nicht</em> gezeigt wird: Ligaturen. iA Writer stammt von IBM Plex ab und kennt keine
          Programmier-Ligaturen, <code class="font-mono text-xs">-&gt; =&gt; != &gt;= ===</code> bleiben also
          Einzelzeichen.
        </p>
      </section>

      <section class="space-y-2">
        <h2 class="text-sm font-semibold text-highlighted">Die Strichstärke ist stufenlos — von 400 bis 700</h2>
        <p class="text-sm text-default">
          Eine Datei, jede Stärke dazwischen. Der Bereich ist aber nicht der übliche von 100 bis 900: diese Dateien
          tragen <code class="font-mono text-xs">wght 400–700</code>, und was darüber hinaus verlangt wird, wird auf
          die Grenze gezogen. Unten sieht man das — 100 bis 400 sind dasselbe, 700 bis 900 auch.
        </p>
        <div class="space-y-1 rounded-sm border border-accented bg-default p-4">
          <p
            v-for="weight in weights"
            :key="weight"
            class="font-quattro text-lg"
            :style="{ fontWeight: weight }"
          >
            {{ weight }} — Handgloves 0123
          </p>
        </div>
      </section>

      <section class="space-y-2">
        <h2 class="text-sm font-semibold text-highlighted">Die zweite Achse: Laufweite</h2>
        <p class="text-sm text-default">
          Neben der Strichstärke tragen die iA-Dateien eine Achse <code class="font-mono text-xs">SPCG</code> von 0
          bis 150, die den Abstand zwischen den Zeichen öffnet. Über Tailwind ist sie nicht erreichbar, über
          <code class="font-mono text-xs">font-variation-settings</code> schon. Benutzt wird sie bisher nirgends.
          Inter hat statt dessen eine Achse <code class="font-mono text-xs">opsz</code>, die den Schnitt an die
          Schriftgröße anpasst.
        </p>
        <div class="space-y-1 rounded-sm border border-accented bg-default p-4">
          <p
            v-for="spacing in spacings"
            :key="spacing"
            class="font-mono text-lg"
            :style="{ fontVariationSettings: `'SPCG' ${spacing}` }"
          >
            SPCG {{ spacing }} — Handgloves 0123
          </p>
        </div>
      </section>

      <section class="space-y-2">
        <h2 class="text-sm font-semibold text-highlighted">Das Fragment</h2>
        <p class="text-sm text-default">
          Genau dieser Ausschnitt erzeugt die Probe oben — er wird gerendert und hier ausgegeben, damit das eine nicht
          vom anderen abweichen kann.
        </p>
        <pre class="overflow-x-auto rounded-sm border border-accented bg-default p-4 font-mono text-xs/relaxed text-default">{{ fragment.trim() }}</pre>
      </section>

      <section class="space-y-2">
        <h2 class="text-sm font-semibold text-highlighted">Inter und die Schriften des Systems</h2>
        <p class="text-sm text-default">
          Das Paket bringt Inter für alle Schriftsysteme mit, ausgeliefert wird davon fast nichts: jede Schnittstelle
          trägt eine <code class="font-mono text-xs">unicode-range</code>, und der Browser holt nur die Datei, deren
          Zeichen er braucht. Für Deutsch ist das die lateinische — die Umlaute stehen dort —, Kyrillisch,
          Griechisch und Vietnamesisch bleiben liegen.
        </p>
        <p class="font-inter text-lg">Inter — Handgloves 0123 (variabel, 100 bis 900)</p>
        <p class="font-inter text-lg" style="font-weight: 200">Inter 200 — Handgloves 0123 · Grüße aus Zürich</p>
        <p class="font-inter text-lg" style="font-weight: 800">Inter 800 — Handgloves 0123 · Grüße aus Zürich</p>
        <p class="text-lg" style="font-family: monospace">Monospace des Systems — Handgloves 0123</p>
        <p class="text-lg" style="font-family: sans-serif">Serifenlose des Systems — Handgloves 0123</p>
      </section>
    </div>
  </article>
</template>

<script setup lang="ts">
// What the project's typefaces are, and how to tell them apart.
//
// A page rather than a note, because a font either renders or it does not: this
// shows the three families actually resolving, and shows the markup that does
// it, so a family that stopped loading is visible here rather than in a
// screenshot somebody took once.

// Past what the files carry at either end on purpose: a ladder is only honest
// if it shows where it stops.
const weights = [100, 300, 400, 500, 600, 700, 900]

const spacings = [0, 50, 100, 150]

// Rendered above and printed below, from this one string.
const fragment = `
<p style="font-family: 'iA Writer Mono'">Mono — Handgloves 0123 · mmmm wwww iiii llll</p>
<p style="font-family: 'iA Writer Mono'; font-weight: 700">Mono fett — Handgloves 0123</p>
<p style="font-family: 'iA Writer Duo'">Duo — Handgloves 0123 · mmmm wwww iiii llll</p>
<p style="font-family: 'iA Writer Quattro'">Quattro — Handgloves 0123 · mmmm wwww iiii llll</p>
<p style="font-family: 'iA Writer Quattro'; font-style: italic">Quattro kursiv — Handgloves 0123</p>
`
</script>
