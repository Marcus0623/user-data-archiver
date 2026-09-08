# User Data Archiver

<p>
  <a href="README.md"><kbd>English</kbd></a>
  &nbsp;<a href="README.zh-CN.md"><kbd>简体中文</kbd></a>
  &nbsp;<a href="README.zh-TW.md"><kbd>繁體中文</kbd></a>
  &nbsp;<a href="README.ja.md"><kbd>日本語</kbd></a>
  &nbsp;<kbd><b>Français</b></kbd>
  &nbsp;<a href="README.ru.md"><kbd>Русский</kbd></a>
  &nbsp;<a href="README.vi.md"><kbd>Tiếng Việt</kbd></a>
</p>

Outil Windows qui archive les **fichiers utilisateur hors OS** vers un autre disque ou une clé USB, **en conservant l’arborescence**. Usage typique : récupérer les fichiers personnels d’un collègue qui part.

Programme Go autonome. Après chaque mise à jour, ce dossier ne doit contenir **qu’un** exécutable : `user-data-archiver.exe`.

## Fonction

Parcourt `C:` (ou d’autres chemins) et enregistre ce qui **ne reviendrait pas** après une réinstallation propre : bureau, documents, téléchargements, projets, messagerie, et en mode recommandé les réglages navigateur. Ignore par défaut l’OS, la Corbeille, le fichier d’échange, les programmes installés, les dossiers OEM et les caches.

## Copier ou couper

Première fenêtre :

| Choix | Effet |
| --- | --- |
| **Copy files** | Enregistre la destination ; les originaux restent |
| **Cut files (move)** | Enregistre, vide les tampons, puis **supprime l’original seulement si la destination est lisible et identique octet par octet**. Taille/date identiques ne suffisent pas. Les fichiers cloud « en ligne seulement » ne sont pas supprimés. Irréversible. Ne supprime pas `C:\`, `C:\Users` ni la racine d’un profil |
| **Cancel** | Quitter |

Le mode couper redemande confirmation (Non par défaut). L’aperçu ne copie ni ne supprime.

## Fenêtre

Double-clic sur `user-data-archiver.exe` ou `start-archive.bat`. Ensuite : nom, source, **Enregistrer vers** (Parcourir ; un dossier sur un disque source est permis, pas la racine `D:\`), **Exclure**, type/libellé/espace libre du disque, **Test lecture/écriture**, mode, options Program Files / `node_modules`, aperçu, démarrer / arrêter, journal. La source accepte plusieurs chemins (virgule ou point-virgule, ex. `C:,D:`). **Parcourir** ajoute sans effacer le champ. Si **Aperçu uniquement** est coché, utilisez **Aperçu**. **Démarrer** copie ou coupe vraiment. Les droits administrateur servent à **lire** les autres profils sur C:, pas à écrire sur la destination. `-cli` ou `-yes` : pas de fenêtre. `-cut` ouvre directement le mode couper. Langue : `-lang` (`en`, `zh-CN`, `zh-TW`, `ja`, `fr`, `ru`, `vi`).

## Exemple

Alice quitte l’entreprise. Windows et son profil sont sur `C:`, les projets sur `D:\Projects`. Clé USB `E:`. Objectif : **copier** (garder les originaux).

1. Ouvrir `user-data-archiver.exe` et choisir **Copier les fichiers**.
2. Nom `alice`. Source `C:,D:\Projects` (ou Parcourir `C:\` puis Parcourir `D:\Projects` — le second ajout s’ajoute).
3. Destination `E:\offboarding-archive\alice`. Test lecture/écriture → **Aperçu** → **Démarrer la copie**.

| Sur le PC | Dans l’archive |
| --- | --- |
| `C:\Users\alice\Desktop\handoff.docx` | `E:\offboarding-archive\alice\C\Users\alice\Desktop\handoff.docx` |
| `C:\Downloads\contract.pdf` | `E:\offboarding-archive\alice\C\Downloads\contract.pdf` |
| `D:\Projects\api\readme.md` | `E:\offboarding-archive\alice\D\Projects\api\readme.md` |

Le dossier contient aussi `_archive-report.txt`. En ligne de commande :

```bat
user-data-archiver.exe -lang fr -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
```

Sans USB :

```bat
user-data-archiver.exe -lang fr -name alice -src C:,D: -dst D:\offboarding-archive\alice -mode appdata -yes
```

Placez la destination sur **un autre disque ou USB** si possible. Sans USB : source `C:,D:`, destination `D:\offboarding-archive\alice` (pas `D:\`). Un avertissement « destination dans la source » est normal ; le dossier d’archive est ignoré pour ne pas se copier en lui-même. D: doit avoir assez d’espace pour une **seconde** copie.

## Messagerie

Le mode **appdata** (recommandé) archive les discussions **déjà présentes sur ce PC**, p. ex. `Documents\WeChat Files` et `AppData\Roaming` (WeChat/QQ, DingTalk, Lark, Telegram, Teams…). Pas (ou incomplet) : messages uniquement dans le cloud ou sur téléphone, AppData en mode **personal**, fichiers verrouillés si l’appli est ouverte, OneDrive « en ligne seulement ». C’est une copie de fichiers, pas un export de conversation lisible.

## Chemins, filtres, reprise

Pas de `D:\C:\Downloads` : la lettre de lecteur devient un dossier. Destination de préférence sur **un autre disque ou USB**. `C:,D:` vers `D:\offboarding-archive\<nom>` est permis (le dossier d’archive est ignoré ; `D:\` racine est refusé). Ignore `Windows`, `Program Files`, caches, `node_modules`, etc. Mode **appdata** : `AppData\Roaming` (messagerie/navigateur) sans caches. `Windows.old\Users` est inclus. Reprise de copie sur la **même destination** si taille et date identiques. Couper ne supprime qu’après comparaison du contenu. Rapport `_archive-report.txt` ; échecs `_failed-files.csv`.

## Compilation et exécutable

Un seul `user-data-archiver.exe` dans ce dossier. `build.bat` supprime les autres `.exe` et n’en garde qu’un. `start-archive.bat` recompile si les sources sont plus récentes.

## Modes et ligne de commande

**personal** / **appdata** (recommandé) / **all**. Déplacer : `-cut`. Aperçu : `-dry-run`.

```bat
user-data-archiver.exe -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -name alice -src C:,D: -dst D:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
user-data-archiver.exe -lang fr
```

| Option | Signification |
| --- | --- |
| `-name` | Nom (rapport et dossier par défaut) |
| `-src` | Sources, séparées par des virgules |
| `-dst` | Dossier de destination (un dossier sur un disque source est permis ; pas `D:\`) |
| `-mode` | `personal` \| `appdata` \| `all` (défaut `appdata`) |
| `-include-program-files` | Inclure aussi Program Files / ProgramData |
| `-include-regeneratable` | Inclure `node_modules`, `__pycache__`, etc. |
| `-exclude` | Noms de dossiers à ignorer |
| `-dry-run` | Scan uniquement |
| `-cut` | Déplacer : supprimer l’original après vérification |
| `-lang` | Langue : `en`, `zh-CN`, `zh-TW`, `ja`, `fr`, `ru`, `vi` |
| `-cli` | Invites en ligne de commande |
| `-yes` | Sans questions ni fenêtre (`-dst` et `-src` requis) |

Usage autorisé uniquement.
