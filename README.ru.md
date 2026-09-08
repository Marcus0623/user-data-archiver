# User Data Archiver

<p>
  <a href="README.md"><kbd>English</kbd></a>
  &nbsp;<a href="README.zh-CN.md"><kbd>简体中文</kbd></a>
  &nbsp;<a href="README.zh-TW.md"><kbd>繁體中文</kbd></a>
  &nbsp;<a href="README.ja.md"><kbd>日本語</kbd></a>
  &nbsp;<a href="README.fr.md"><kbd>Français</kbd></a>
  &nbsp;<kbd><b>Русский</b></kbd>
  &nbsp;<a href="README.vi.md"><kbd>Tiếng Việt</kbd></a>
</p>

Инструмент Windows: архивирует **пользовательские файлы вне ОС** на другой диск или USB и **сохраняет исходное дерево папок**. Типичное применение — сбор личных файлов увольняющегося коллеги.

Отдельная программа на Go. После каждого обновления в папке должен оставаться **один** файл: `user-data-archiver.exe`.

## Назначение

Обходит `C:` и сохраняет то, **чего не будет после чистой переустановки**: рабочий стол, документы, загрузки, проекты, мессенджеры и в рекомендуемом режиме настройки браузера. По умолчанию пропускаются ОС, Корзина, файл подкачки, установленные программы, OEM и кэши.

## Копирование или вырезание

Первое окно:

| Выбор | Действие |
| --- | --- |
| **Copy files** | Сохранить копию; оригиналы остаются |
| **Cut files (move)** | Сохранить, сбросить на диск, затем **удалить оригинал только если назначение читается и побайтно совпадает**. Совпадения размера и времени недостаточно. Облачные заглушки не удаляются. Необратимо. Не удаляет `C:\`, `C:\Users` и корень профиля |
| **Cancel** | Выход |

Перед вырезанием — повторное подтверждение (по умолчанию Нет). Просмотр ничего не копирует и не удаляет.

## Окно

Двойной щелчок по `user-data-archiver.exe` или `start-archive.bat`. Далее: имя, источник, **Сохранить в** (папка на исходном диске разрешена, корень `D:\` — нет), **Исключить**, тип/метка/свободное место диска, **Тест чтения/записи**, режим, Program Files / `node_modules`, просмотр, старт / стоп, журнал. Источник: несколько путей через запятую или точку с запятой (`C:,D:`). **Обзор** добавляет путь, не затирая поле. Если включён **только предпросмотр**, используйте **Предпросмотр**. **Старт** всегда копирует или вырезает. Права администратора нужны, чтобы **читать** чужие профили на C:, а не чтобы писать на диск назначения. `-cli` / `-yes` — без окон. `-cut` сразу открывает режим вырезания. Язык: `-lang` (`en`, `zh-CN`, `zh-TW`, `ja`, `fr`, `ru`, `vi`).

## Пример

Alice увольняется. Система и профиль на `C:`, проекты на `D:\Projects`. USB — `E:`. Нужно **копирование** (оригиналы остаются).

1. Откройте `user-data-archiver.exe` и выберите **Копировать файлы**.
2. Имя `alice`. Источник `C:,D:\Projects` (или Обзор `C:\`, затем Обзор `D:\Projects` — второй путь добавится).
3. Назначение `E:\offboarding-archive\alice`. Тест чтения/записи → **Предпросмотр** → **Начать копирование**.

| На ПК | В архиве |
| --- | --- |
| `C:\Users\alice\Desktop\handoff.docx` | `E:\offboarding-archive\alice\C\Users\alice\Desktop\handoff.docx` |
| `C:\Downloads\contract.pdf` | `E:\offboarding-archive\alice\C\Downloads\contract.pdf` |
| `D:\Projects\api\readme.md` | `E:\offboarding-archive\alice\D\Projects\api\readme.md` |

В той же папке — `_archive-report.txt`. Командная строка:

```bat
user-data-archiver.exe -lang ru -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
```

Без USB:

```bat
user-data-archiver.exe -lang ru -name alice -src C:,D: -dst D:\offboarding-archive\alice -mode appdata -yes
```

Пишите на **другой диск или USB**, если можно. Без USB: источник `C:,D:`, назначение `D:\offboarding-archive\alice` (не `D:\`). Предупреждение «назначение внутри источника» — нормально; папка архива пропускается, чтобы не копировать саму себя. На D: должно хватить места на **вторую** копию.

## Чаты

В режиме **appdata** копируются переписки, **уже лежащие на этом ПК**: например `Documents\WeChat Files` и `AppData\Roaming` (WeChat/QQ, DingTalk, Lark, Telegram, Teams…). Нет или неполно: только облако/телефон, AppData в режиме **personal**, файлы, открытые в приложении, заглушки OneDrive. Это резервная копия файлов, не читаемый экспорт чата.

## Пути, исключения, докачка

Нельзя `D:\C:\Downloads`: буква диска становится папкой. Лучше писать на **другой диск или USB**. Можно сканировать `C:,D:` и сохранять в `D:\offboarding-archive\<имя>` (папка архива пропускается; корень `D:\` запрещён). Пропускаются `Windows`, `Program Files`, кэши, `node_modules`. В **appdata** копируется `AppData\Roaming` (мессенджеры/браузер) без кэша. `Windows.old\Users` включается. Повтор **копирования** на ту же папку пропускает файлы с тем же размером и временем. Вырезание удаляет только после сверки содержимого. Отчёт `_archive-report.txt`, ошибки `_failed-files.csv`.

## Сборка и exe

В папке должен быть **один** `user-data-archiver.exe`. `build.bat` удаляет лишние `.exe` и оставляет свежую сборку. `start-archive.bat` пересобирает, если исходники новее.

## Режимы и CLI

**personal** / **appdata** (рекомендуется) / **all**. Перенос: `-cut`. Просмотр: `-dry-run`.

```bat
user-data-archiver.exe -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -name alice -src C:,D: -dst D:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
user-data-archiver.exe -lang ru
```

| Флаг | Значение |
| --- | --- |
| `-name` | Имя (отчёт и папка по умолчанию) |
| `-src` | Источники через запятую |
| `-dst` | Папка назначения (на исходном диске можно; корень `D:\` нельзя) |
| `-mode` | `personal` \| `appdata` \| `all` (по умолчанию `appdata`) |
| `-include-program-files` | Включить Program Files / ProgramData |
| `-include-regeneratable` | Включить `node_modules`, `__pycache__` и т.п. |
| `-exclude` | Дополнительно пропускать эти имена папок |
| `-dry-run` | Только сканирование |
| `-cut` | Перенос: удалять оригинал после проверки назначения |
| `-lang` | Язык: `en`, `zh-CN`, `zh-TW`, `ja`, `fr`, `ru`, `vi` |
| `-cli` | Вопросы в командной строке |
| `-yes` | Без вопросов и окон (нужны `-dst` и `-src`) |

Только для разрешённой передачи дел или своей копии.
