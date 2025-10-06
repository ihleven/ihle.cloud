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
