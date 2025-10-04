# Update

## Handler

- ermittelt Account/User
- parst entry aus body und checkt dabei syntaktische Integrität (ob keine Daten 'verloren' gehen, optional)
- übergibt dann an Usecase und bekommt Changeset zurück
- je nach Modus (param `mode`) wird das Changeset gespeichert -> ?besser in Usecase?
- Rückgabe Changeset

## Usecase

Hat Zugriff auf Repo, liest daraus die alten Daten und speichert die neuen

- bekommt die neuen Daten und den verantwortlichen User übergeben, evtl. auch Modus (dryrun?)
- IsValidUpdateTo 


## Legacy

`cms.Update(entry, account, msg, repo)`
repo.Get
cms.CheckConstraints
entry.IsValidUpdateTo
Version handling 
commit