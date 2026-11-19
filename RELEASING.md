# Releasing GoPlay

Publishing a GitHub release with a tag like `v1.2.0` runs
`.github/workflows/build.yml`. The tag sets the version, and the run number
sets the build number (`CFBundleVersion`), so every run uploads a higher one.
The workflow tests on each runner, builds every package and attaches them to
the release. A manual run with a tag rebuilds that release; without a tag it
keeps the packages as the run's artifacts.

The workflow is written by hand, not by `giode ci`: libmpv comes from Homebrew,
apt and shinchiro's Windows builds, and the tests need generated media. Its
store steps follow `giode ci`'s, with the same secret names.

| Where | What the release gets | Needs |
| --- | --- | --- |
| GitHub, Linux | `.deb`, `.rpm` and `.tar.gz`, x64 and ARM64 | nothing |
| GitHub, Windows | `-setup.exe` installers and portable `.zip`s, x64 and ARM64 | nothing; a code signing certificate would stop SmartScreen warnings (not set up) |
| GitHub, macOS | `GoPlay-<version>-macos-arm64.dmg` (Apple silicon) | a Developer ID certificate, or macOS refuses to open it; the notary secrets to notarize it |
| Mac App Store | `GoPlay-<version>-macos-arm64-appstore.pkg`, to upload with Transporter | an Apple Developer account, the store certificates and profile |
| Microsoft Store | `GoPlay-<version>-windows-amd64.msix` and `-arm64.msix`, to upload in Partner Center | a Partner Center account and the reserved name |

Not in a store:

- **Google Play and the iOS App Store.** GoPlay is a desktop app
  (`app.platforms` is macOS, Windows and Linux): it plays through libmpv with
  cgo and draws video with mpv's OpenGL render API under Giode's underlay.
  A phone version would need a different playback engine.
- **Snap Store and Flathub.** Linux users download from the GitHub release.

## Repository secrets

Set these in GitHub under Settings → Secrets and variables → Actions. Each
`*_BASE64` secret is the file encoded with `base64 -i file | pbcopy` on macOS,
or `base64 -w0 file` on Linux. Without a secret, that step is skipped and the
rest of the release still builds (an unsigned `.dmg`, no `.pkg`).

| Secret | What it is |
| --- | --- |
| `MACOS_CERT_BASE64`, `MACOS_CERT_PASS` | The "Developer ID Application" certificate as a `.p12`, and its password: signs the `.dmg` |
| `GIODE_NOTARY_ID`, `GIODE_NOTARY_TEAM_ID`, `GIODE_NOTARY_PASS` | Your Apple ID, team ID and an app-specific password: notarizes the `.dmg` |
| `MACOS_STORE_CERT_BASE64`, `MACOS_STORE_CERT_PASS` | The "Apple Distribution" certificate as a `.p12` |
| `MACOS_INSTALLER_CERT_BASE64`, `MACOS_INSTALLER_CERT_PASS` | The "Mac Installer Distribution" certificate as a `.p12` |
| `MACOS_PROFILE_BASE64` | The Mac App Store provisioning profile (`.provisionprofile`) |

The Microsoft Store signs the `.msix` itself, so Windows needs no secret.

## Mac App Store

1. Join the Apple Developer Program.
2. In Certificates, Identifiers & Profiles:
   - Register the App ID `com.mehran9675.goplay` (the `app.id` in
     `giode.toml`), with no extra capabilities.
   - Create an "Apple Distribution" and a "Mac Installer Distribution"
     certificate, and a "Developer ID Application" one for the `.dmg`.
     Export each from Keychain Access as a `.p12`.
   - Create a "Mac App Store" provisioning profile for the App ID.
3. Create an app-specific password at account.apple.com for notarization.
4. In App Store Connect, create a macOS app with the bundle ID
   `com.mehran9675.goplay`, SKU `goplay`.
5. Add the secrets above, then publish a release. Download
   `GoPlay-<version>-macos-arm64-appstore.pkg` from it and upload it with the
   Transporter app.
6. Fill in the listing with the copy below; set the privacy policy URL to
   https://xmhrnx.github.io/goplay/#privacy.

**Before the first submission, check the minimum macOS.** `giode.toml` says
macOS 11, but the CI bundles Homebrew's libmpv and FFmpeg from the
`macos-latest` runner, which are built for that runner's macOS. Run
`otool -l GoPlay.app/Contents/Frameworks/*.dylib | grep -A3 LC_BUILD_VERSION`
on the built app and raise `[macos] min_os` to the highest `minos` it shows,
or older Macs would get a GoPlay that doesn't start. The package is Apple
silicon only: Intel Macs can't install it.

**The App Sandbox limits GoPlay in the store version.** Every Mac App Store
app runs sandboxed. `[macos.store]` in `giode.toml` gives GoPlay the network
(as a client and a server, for casting), read access to what the user opens,
and Pictures for screenshots. So in the store version:

- **yt-dlp and FFmpeg can't run.** Web pages (YouTube and other sites) don't
  play; direct links and HLS/DASH streams do. Casting a file a TV can't play
  doesn't convert it; files the TV plays cast as usual.
- **Only what the user opens is readable.** Opening one episode doesn't queue
  the rest of its folder, and subtitles next to a video aren't loaded by
  themselves: open the folder with Open Folder, or drop the subtitle file on
  the window. Folders opened with Open Folder are remembered across
  restarts; single files in the recent list may need opening again.
- **Discs** play from a mounted disc or a disc folder chosen in Open DVD or
  Blu-ray, not straight from the drive's device.
- **Open Config Folder, Edit mpv.conf and Show in Finder** run `open`, which
  the sandbox may refuse; settings are in the app's container,
  `~/Library/Containers/com.mehran9675.goplay/Data/Library/Application Support/goplay`.

The notarized `.dmg` on GitHub has none of these limits. Point users who need
them to it.

## Microsoft Store

1. Open a Partner Center developer account (the same one as EasySend and
   AnyFile).
2. Reserve the name "GoPlay". If it's taken, reserve another and see the note
   below.
3. In the product's Product identity page, check these against `giode.toml`
   and fix any that differ:
   - **Package/Identity/Name** → `identity_name` under `[windows.msix]`
     (now `MehranZahir.GoPlay`).
   - **Package/Identity/Publisher** → `publisher` under `[windows.msix]`
     (`CN=B40E8F57-A7E0-40BB-8E74-5D051D0B578B`).
   - **Package/Properties/PublisherDisplayName** → `publisher_display_name`
     under `[windows.msix]` (`Mehran Zahiroddini`).
4. Commit and publish a release. Upload both
   `GoPlay-<version>-windows-amd64.msix` and `-arm64.msix` in one submission.
5. When the submission asks why GoPlay uses the `runFullTrust` capability:
   GoPlay is a desktop video player that loads libmpv and its codecs, opens
   any file the user chooses, and runs a local HTTP server so TVs on the home
   network can stream the file being cast. Its manifest declares an inbound
   firewall rule for that server (`firewall = true`).
6. Privacy policy URL: https://xmhrnx.github.io/goplay/#privacy.

If the reserved name isn't "GoPlay", the MSIX's display name must match it:
set `app.name` to the reserved name for the MSIX build only, as AnyFile's
workflow does.

## Store listing

Accurate to the code as of this release. Screenshots and claims should match
what the store version can do (see the sandbox list above for the Mac).

- **Name:** GoPlay
- **Subtitle (Apple, ≤30):** Plays almost anything
- **Short description (≤80):** A fast video player built on mpv: subtitles, HDR, 3D and 360° video, casting.
- **Category:** Photo & Video (both stores). The Mac build declares
  `public.app-category.video`.
- **Keywords (Apple, ≤100):**
  `video player,mpv,mkv,mp4,subtitles,hdr,3d,vr,360,chromecast,dlna,cast,movies,media player,hevc,av1`
- **Support URL:** https://github.com/xmhrnx/goplay/issues
- **Marketing URL:** https://xmhrnx.github.io/goplay/
- **Privacy policy URL:** https://xmhrnx.github.io/goplay/#privacy

**Full description:**

> GoPlay is a video player for people who just want to watch something, with
> depth for those who want to tune everything. It plays nearly every video and
> audio format with mpv's engine, with hardware decoding and video drawn
> straight from the GPU.
>
> Watching
> • Plays nearly every video and audio format, from a file, a folder or a link, including HLS and DASH streams.
> • Resumes where you left off, with your recent files on the start screen.
> • Opening a folder queues it in natural order (episode 2 before 10).
> • DVD and Blu-ray folders and disc images, starting at the main movie.
> • Browse and play videos and music shared by DLNA/UPnP media servers on your network.
>
> Controls
> • Controls sit above and below the picture in a window, and float over it in fullscreen.
> • Timeline with chapters, buffered ranges and an A–B loop.
> • Speed from 0.25× to 4× with pitch kept, frame stepping, repeat and shuffle.
> • Playlist and chapters panel, picture-in-picture and always on top.
> • A command palette for every command, and every shortcut can be changed.
> • In 53 languages.
>
> Picture and sound
> • Video, audio and subtitle tracks, a second subtitle line, external subtitle and audio files.
> • Subtitle size, position, color, outline, background, delay and encoding.
> • Aspect ratio, zoom, rotation, deinterlacing, debanding, HDR tone mapping and color adjustments.
> • Equalizer presets, loudness normalization, night mode and audio delay.
>
> 3D and VR
> • Side-by-side and top-and-bottom 3D as anaglyph, one eye, interleaved or for 3D TVs.
> • 360° and VR180 video: drag to look around, scroll to zoom.
>
> Casting
> • Cast to Chromecast, Google TV, Android TV and DLNA/UPnP TVs and receivers on your network.
>
> For power users
> • Any mpv option, mpv.conf, input.conf and Lua or JavaScript scripts.
>
> No accounts, ads, analytics or tracking.

For the Microsoft Store, the same text works; it can add "Web pages from
YouTube and other sites play when yt-dlp is installed, and casting converts
files a TV can't play when FFmpeg is installed." Leave that sentence out of
the Mac App Store description.

**Screenshots.** None are in the repository yet. Take them from the store
build at 2880×1800 (Mac, 16:10) and 1920×1080 or larger (Windows), with
footage you may publish (Blender's open movies, such as Big Buck Bunny,
credited as their licence asks):

1. The start screen with recent files.
2. A movie playing in a window, with the timeline and chapters.
3. The playlist and chapters panel.
4. Subtitle settings, or the subtitle menu with two tracks.
5. The cast dialog listing devices.
6. A 360° video, or the 3D menu.

**Age rating.** No violence, sexual content, gambling, user-generated
content, chat or purchases: Apple 4+, IARC 3/Everyone. GoPlay has no web
browser, but it plays any link the user types; if the questionnaire asks about
unrestricted web access, answer "No" and, if review asks, explain that it
plays media links the user enters and has no browser (answer "Yes" if they
disagree; that gives the highest age rating on Apple).

**App Privacy (Apple) and data safety (Microsoft):** Data Not Collected.
GoPlay has no accounts, analytics, crash reporting or tracking, and sends
nothing to its developer; settings and the recent list stay on the computer.
`[privacy]` in `giode.toml` builds the matching `PrivacyInfo.xcprivacy`
(no tracking, nothing collected).

**Export compliance:** answered in the build
(`ITSAppUsesNonExemptEncryption = false`): GoPlay only uses HTTPS and other
standard protocols, through macOS and libmpv/FFmpeg, to play streams.

## Licence

The repository has no LICENSE file, while `giode.toml`'s `[linux] license`
says MIT and the `.deb`/`.rpm` carry that. Add the licence file you intend
(and keep `giode.toml` in step) before submitting: the stores' listings and
Linux packaging expect one. GoPlay also bundles libmpv and FFmpeg, and the
builds the CI uses (Homebrew's, shinchiro's) are GPL-licensed: pick a licence
compatible with that, ship their notices, and offer their source. The GPL's
fit with the App Store's terms has been questioned before; decide on that
before submitting to the Mac App Store.

## Linux

The `.deb` and `.rpm` depend on the system's libmpv (`libmpv2 | libmpv1`,
`mpv-libs`), so `apt` and `dnf` install it. The `.tar.gz` holds the program
for distributions without those packages; it needs libmpv installed.
