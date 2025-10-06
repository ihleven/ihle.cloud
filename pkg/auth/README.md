




account mit 
 * username
 * pwdhash
 * hidrive.user oder generische hidrive auth 



# Endpunkte 

### Signin

Endpukt für Authentifizierung von Usern. Parameter sind usersna eund passwort.
Wenn ok, dann wird:
1. jwt erzeugt und als cookie (jwt) dem Frontend übergeben.
2. geschaut, ob für den hinterlegten hi account ein valides Token verfügbar ist, wenn nicht, dann wird hidrive login prozeess gestartet.

### AuthorizeCallback

behandelt den Callback von HIdrive nach erfolgreichem login

### Info

benutzt um im Frontend die Session daten zu laden (dazu muss das cookie gesetzt sein.)




# Authorization without authentication

* mit query-Param wie token=foo oder auth=bar
* steht für kombination aus hidrive-Token und erlaubten Pfaden
* auth-Middleware schickt param an info-Endpunkte wie auch evt. vorhandenes auth-Cookie
* kann evtl. mehrere authorization-Tokens (aus Cookie) schicken
* authorization-Token über separten login screen für nicht frei zugängliche inhalte