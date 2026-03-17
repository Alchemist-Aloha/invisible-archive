[Setup]
AppName=Invisible Archive
AppVersion=1.0.0
DefaultDirName={autopf}\InvisibleArchive
DefaultGroupName=Invisible Archive
OutputDir=.
OutputBaseFilename=InvisibleArchiveInstaller
Compression=lzma
SolidCompression=yes
SetupIconFile=runner\resources\app_icon.ico

[Files]
Source: "..\build\windows\x64\runner\Release\invisible_archive_mobile.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\windows\x64\runner\Release\*.dll"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\windows\x64\runner\Release\data\*"; DestDir: "{app}\data"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\Invisible Archive"; Filename: "{app}\invisible_archive_mobile.exe"
Name: "{commondesktop}\Invisible Archive"; Filename: "{app}\invisible_archive_mobile.exe"

[Run]
Filename: "{app}\invisible_archive_mobile.exe"; Description: "{cm:LaunchProgram,Invisible Archive}"; Flags: nowait postinstall skipifsilent
