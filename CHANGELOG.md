# CHANGELOG

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

This file starts here; anything before the entries below is only in the git history.

## Unreleased

### Added

- **The files on HiDrive can be browsed in the app**: folders and their contents, with the size, dimensions and date of each entry, a trail back up, and the pictures of a folder shown as pictures rather than as a list of filenames. A single file can be looked at and downloaded. It is its own area, so it appears only for an account that has been given it — having storage configured is not by itself permission to go through it, and an account that has been given the area but has no storage of its own is shown the family's.


- **The football pool is served from here**: Geheimtipp — the tipping pool that has lived at ihleven.de — is now rendered by this app: the current matchdays, a matchday's tip matrix with every tipper side by side, the day's and the overall standings, the league table, the list of matchdays, and the ranking with everyone's avatar and motto. The tips themselves still live in the pool's own database; only the pages are new.

- **Tips can be placed again**: on a matchday, your own column is typeable until each match kicks off — every other column stays as it is, and a match whose kick-off has passed stops taking anything. One *Abschicken* sends the whole matchday at once, and the pool answers per match, so tips it takes are kept even when one of them comes too late. A single match can also be tipped on its own page, with the same up-and-down buttons the pool has always had.

- **Results are entered in the app**: on a matchday, whoever keeps the pool's results gets a field per fixture instead of the printed score, and one button saves the lot — which rescores everyone, so the standings and the league table beside it are re-read with it. Fixtures not yet played keep their placeholder rather than being cleared. Nobody else is offered the fields, and the pool refuses the save from anyone else regardless.

- **A page for each match**: the two clubs over a photograph with the result between them, arrows to the match before and after, and underneath it every tipper's tip with the points it earned, best first.

- **Registering for the edition**: name, motto and a picture — chosen from the pool's collection or uploaded. Signing in and being registered are two different things, and trying to tip without the second now says so and offers the way to fix it, instead of being refused by the server without explanation.

- **A changed name, motto or picture shows up straight away**: previously the registration page saved, and the ranking and the matchday headings went on showing the old one until the page was reloaded — the same thing happened on the old site. A tipper is described in two places at once, and only one of them was being re-read.

- **The avatar page shows tiles or a list**, whichever is wanted: the tiles for picking a picture out, the list for reading what a picture actually is — who it belongs to, its size, format and when it arrived. The old site had only the list.

- **Two menus, one per thing they belong to**: the header over the grass carries the account — who is signed in, and the way out — and stays the same on every page of the pool whatever edition is being read. The bar below carries the edition: its pages, its name, and the registration this person plays it under, with their picture, name, motto and the way to change them. Signing out is offered once, where the account is, rather than in both.

- **The pool has its own sign-in**: whoever tips signs in with their Geheimtipp account, not with the one for the family site. The two live at the same address and have nothing else in common: an account here grants nothing there, and the other way round. Someone signed in sees their own column first — on a phone it is the only one shown, because nobody can read twelve columns of scores on a phone anyway.

- **A new account can be set up without a terminal**: creating one, giving it a first password and handing over an invitation link are all in the app now. Registering a device still asks for that password, as it always has — the link says who, the password proves it is them, and the two are meant to reach the person by different routes.

- **Accounts can be administered from the app**: who may sign in, what each person gets to see, and what they sign in with — all of it now in a "Konten" section instead of only at a terminal on the machine holding the database. An account is created, given its areas, and handed an invitation link; it can be locked without being deleted, signed out everywhere, or have a lost device removed. The section only appears for an account that has been given it, and the server refuses the same requests it hides, so the two cannot disagree. Setting up the very first administrator is still a command-line job, as is any situation where the app itself is what is broken.

- **The family films are a kind of content in their own right**: a film is now described the way a film actually is — which format it was shot on, when and where and by whom, how long it runs, a description, and a still to represent it. It replaces a description built around one particular storage path, so the archive can hold an 8mm reel or a video tape later without anything being redesigned.

- **Films are divided into named scenes**: a reel is a series of runs of the camera, and each one can now be given a name, a starting point and a description. The player offers them as chapters, using the browser's own chapter support rather than a hand-built imitation of it, so jumping to "Bescherung" works the way chapters work everywhere else. Only the starting point has to be given: a scene runs until the next one begins.

- **People and places can be marked in a film, at the moment they appear**: an annotation says that someone is visible from here to there, and optionally *where* on the picture, so a name can be shown next to the person it belongs to. Positions are recorded as a fraction of the frame rather than in pixels, so they stay correct however large the video is displayed. An annotation lasts five seconds unless an end is given. Names are free text — a contributor can write a name nobody has an entry for, or an uncertain one — and they are searchable, so "which films is Oma Anna in" is a question the archive can answer. Each annotation records who added it, in preparation for letting the family add them.

- **An editor for films**: the film, its scenes and its annotations are edited beside a video player with a proper transport — play and pause, and steps of half a second, a tenth and a hundredth, with the exact time shown to the millisecond, so a scene boundary can be placed precisely. A time can be taken from wherever the video is paused, any time in the lists jumps the video there, and an annotation is placed on the picture by clicking it. The same controls are on the keyboard.

- **A page showing all the films**: a list, or a wall of stills in one of three shapes — 4:3, square filled, or square with the whole frame fitted inside. The background can be light or dark. A film with no chosen still is represented by a frame taken from the film itself, so the wall looks like something without anyone having had to pick pictures.

- **Each account sees only the areas it has been given**: the parts of the site — films, content, family, calendar, media library, music, search — are now granted per account rather than being the same for everyone. The navigation, the start page and the links all follow what an account actually has, so someone given only the films sees only the films, and a visitor who is not signed in sees nothing. This hides areas rather than sealing them off; it is about not offering people doors that are not theirs.

- **The app can be added to an iPhone's home screen**: opened from there it runs without the browser's address bar and toolbar, with its own name and icon. Nothing about how updates reach people changes — a new version is picked up the moment it is deployed, as before.

- **A command for checking the storage connection**: `ihlvn hitoken` reports which storage accounts the app holds credentials for, whether each one still works and when it expires — so a failure to play a film can be told apart from a failure to reach the storage provider without guessing.

### Changed

- **A film is addressed by which film it is, not by where its file sits**: previously the page was handed the path of the video inside the storage provider and asked for it by that path, which put the storage layout into the address bar and into the page's source. Now the page asks for the film, and the server looks up where the bytes are. Nothing about the storage arrangement reaches the browser, and moving a file is a change to the film's entry rather than to every link that pointed at it. Measured on the same film, it is also faster to start playing — the previous route asked the storage provider for a fresh address on every single request, including every jump within a video, and that round trip is gone.

- **Signing in and out moved out of the top bar and into the menu**, which is now reachable at every window size rather than only on a phone, and is split into where you can go on one side and who you are on the other. The account details are shown directly instead of hiding behind a hover menu.

- **The site's name appears in the footer on every page** and, in the top bar, gives way to a way back when you are watching a single film.

### Fixed

- **An account granted everything was reported as being granted nothing**: `*` means "all rights", and the admin commands listed it as a right this build does not know, warning that it would grant nothing. It grants everything, including every area of the site. A right that genuinely is unknown is still reported.

- **Signing in with a passkey failed in some browsers** with "no ceremony is in progress". A browser can have two sign-in attempts open at once — one offered quietly in the username field, one started by the button — and the app could only remember one of them, so whichever finished second found nothing waiting for it. It also forgot the attempt before checking it, so a single mistyped or cancelled attempt broke the next one. The page now says which attempt it is completing, and the app only forgets an attempt once it has succeeded. Safari happened not to hit this; Chrome did.

- **The app no longer hammers the storage provider with expired credentials**: when refreshing its access to the storage provider failed, it retried immediately and indefinitely, and could throw away the long-lived credential by overwriting it with an empty one from a failed response. The provider noticed the traffic. It now waits longer between attempts, keeps the long-lived credential when a response does not repeat it, and stops entirely when the provider says the credential is no longer valid — which is a situation a human has to resolve, not something retrying can fix.

- **Captions in the film player never actually appeared**: every caption in every film was stored without any timing, and the player was reading those timings in a format they were never written in, so it produced captions that began and ended at no time at all. The timings existed all along in a second, separate field. Scenes now carry them, and the player uses the browser's own mechanism to display them.

- **The still shown for a film in listings was pointing at an address that no longer existed**, so film listings showed broken images. Listings now ask for the film's own still.

- **Previews in the film list were stretched**: the preview box is wider than it is tall in a different proportion than the pictures, and nothing told the picture how to fit, so every one was squashed.

### Removed

- **The browser is no longer given a key to the storage provider** (security): on signing in, and on every session check, the app handed the browser an access token for the cloud storage account, in a cookie. Nothing used it — the feature it had been added for was gone — so it was an unnecessary copy of a credential sitting in every visitor's browser. It is gone, and with it the sign-in code's knowledge of the storage provider entirely.

- **The old Super 8 content type**, together with its editor, its preview card and its page, replaced by the film described above.
