---
title: Testseite
prefix: toBeDefinedLater
---

# ihle.cloud

Unter api.ihle.cloud oder ihle.cloud/(content-)api soll die zentrale api für verschiedene Frontends liegen:
- texte
- geheimtipp archiv
- art 
- ...

Frontend mit `yarn dev` und Backend mit `make run` starten, dann folgende Seite aufrufen: http://localhost:3000/tmp/ihle.cloud/README.md

# Backend 

## Architektur

### pkg
* api als pkg für einen server/router der func(w http.ResponseWriter, r *http.Request) error Handler unterstützt (dep target)
* auth als pkg für alles mit Auth
* hi als pkg für Zugriff auf Hidrive, nur generische Handler mit access_token -> handler mit unserer auth gehören in anderes Paket
* spa als pkg für Fileserver

## app 
spzifische Anwendungslogik. Anwendungen in pakete gekapselt, die auf alles in pkg direkt zugreifen dürfen, aber auf Infrastruktur (z.b. pkg db) nur per IF.

### app/db 
* implementiert Datenbankzugriff
* darf auf pakete aus pkg () zugreifen, erfüllt selbst aber lediglich if{} aus anderen Paketen und wird deshalb nur von main referenziert/gewired




## cli

* alte api, wird aktuell  durch backend/pkg ersetzt
* main in pkg cld
* altes hi pkg, noch nicht gelöscht weil z.b. noch readseeker impl, evtl.  übernehmen in backend/hi
  * schlecht: type Meta mit accesstoken, readindex und data bytes

# Token-basierte Auth

Frontend: Account- und Tokenbasierte Anmeldung => sorgt für Cookie innerhalb der Nodeapp


::image{src='../../Pictures/wallpapers/Bamboo.jpg' alt='alt image'} 
Das ist ein **Bambusgarten**. 
::

Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. 

::image-grid{caption='Pushing **up** daisies...'}
:::image{src='frontend/public/Daisies.jpg'}

:::


::

Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.   

::container-screen-width{.px-4}
Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui **blandit** praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat.   
::

# Domains


# equiluz.at	13.01.2026	Externe Nameserver »	
Verwalten
# ihle-tschabrun.de	26.01.2026	Subdomains | E-Mail | DNS	
Verwalten
# ihle.cloud	05.09.2026	Externe Nameserver »	
Verwalten
# ihle.fm	16.01.2026	Subdomains | E-Mail | DNS	
Verwalten
# ihleven.de	18.02.2026	Externe Nameserver »	
Verwalten
# tschabrun.de	28.08.2026	Externe Nameserver »	
Verwalten
# tschabrun.eu	31.08.2026	Externe Nameserver »	
Verwalten
# tthias-ihle.de	16.02.2026	Externe Nameserver »	
Verwalten
# wolfgang-ihle.de	17.02.2026	Externe Nameserver »	
Verwalten

# Github-Projekte

## ihleven

### ihle.cloud 
-> dieses Projekt
- sollte in ihlvn umbenannt werden

### drive 
ganz alt, integrieren in ihlvn, dann löschen
war wohl erste version
alte arbeit funktionalität war hier, 2021 -> ihle-cloud

### ihle-cloud 
arbeit aus drive im November 2021 übernommen

### art (gelöscht)
- 3 commits 2024 + 2025, war ein Versuch das Kunstarchiv und die Famlienapp mit Passkeys neu aufzusetzen
- In Ihlvn integriert und gelöscht am 20.9.2026

### wolfgang-ihle 
- altes Kunst archiv
- das was auf tschabrun.eu installiert ist
- akuteller Branch: go-cli

## Geheimtipp/Tippspile

### geheimtipp
- Geheimtipp nuxt3 frontend
- bereits in ihleven integriert

### goheimtipp 
Private Geheimtipp cli

 
## Webseiten
- wolfgangihle: www.wolfgang-ihle.de
- tierklinik-nuxt: tierklinik-tschabrun.com
- equiluz: equiluz.at

## CMS Content  
- cms-content: Private Inhalte
- cms-content-test test repo für CMS

## Mirrors
- cms Privat
- webcc-tool-api Private
- webcc-elastic Private
- webcc-api Private
- webcc-conf Private
- hhd-web Private ColdFusion Updated on Mar 6, 2020

### Latex
- memoiren Private TeX 
- monochrome-malerei PrivateRich Text Format Updated on Nov 6, 2025







### pkg 
pkg's for gophers
-> keine Abhängigkeit mehr, seit cli/hi weg ist. Löschen bringt trotzdem nichts: der Go-Modul-Proxy hält v0.4.0-alpha ohnehin dauerhaft, und wertvoll ist nur die Historie.
- main (default): errors, hidrive, httpauth, httpsrvr, log, routr, templates — daher kommt der Tag
- develop: abgezweigte Nebenlinie, am 27.9.2026 ausgeräumt
  - kalender -> pkg/kalender (mit Tests)
  - arbeit -> app/arbeit; Storage und Handler ungebaut daneben in _repository/_handler, schema.sql dabei
  - llist -> pkg/llist
  - auth, session: kompilierten beide nicht, wertlos. Ideen notiert (Zugang per Code statt Konto, Bearer neben Cookie, Login ohne SPA)
  - legacy/cloud11-api/drive: von pkg/hi überholt; mitgenommen nur die Bildunterschrift-Idee (Caption/Cutline)
### errors Public
-> v0.1.0, ging mit cli/hi. Keine Abhängigkeit mehr.


### cloud11-webapp 
Private
-> dekomissioniert

### cloud11-api 
Private
-> dekomissioniert

### wi Private
???






- presentations Private

- impress.js Public

