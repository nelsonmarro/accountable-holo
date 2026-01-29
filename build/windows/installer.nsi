; Verith Installer Script
; Professional distribution for Windows with Automated PostgreSQL

!include "MUI2.nsh"
!include "FileFunc.nsh"

; --- General Settings ---
Name "Verith"
OutFile "../../dist/Verith_Setup.exe"
InstallDir "$PROGRAMFILES64\Verith"
InstallDirRegKey HKLM "Software\Verith" "Install_Dir"
RequestExecutionLevel admin

; --- UI Settings ---
!define MUI_ABORTWARNING
; !define MUI_ICON"../../assets/logo.png" 
; !define MUI_UNICON"${NSISDIR}\Contrib\Graphics\Icons\modern-uninstall.ico"

; --- Pages ---
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "../../LICENSE.md"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; --- Languages ---
!insertmacro MUI_LANGUAGE "Spanish"

; --- Variables ---
Var DataDir

Function .onInit
    ; Configurar contexto para 'All Users' (Machine)
    SetShellVarContext all
    StrCpy $DataDir "$APPDATA\Verith"
FunctionEnd

; --- Installation Steps ---
Section "Verith Application" SecApp
    SectionIn RO
    SetOutPath "$INSTDIR"

    DetailPrint "Verificando instalaciones previas..."
    ; Detener el servicio si ya existe para liberar archivos
    nsExec::ExecToLog 'net stop VerithDB'
    
    ; 1. Archivos de la Aplicacion (Binarios -> Program Files)
    File "../../dist/windows/Verith.exe"
    
    SetOutPath "$INSTDIR\assets"
    File /r "../../dist/windows/assets\*.*"
    
    ; 2. Datos y Configuracion (Datos -> ProgramData)
    CreateDirectory "$DataDir"
    CreateDirectory "$DataDir\config"
    CreateDirectory "$DataDir\data"
    CreateDirectory "$DataDir\logs"

    ; Copiar configuracion inicial SOLO SI NO EXISTE
    IfFileExists "$DataDir\config\config.yaml" skip_config
        SetOutPath "$DataDir\config"
        File "../../dist/windows/config\config.yaml"
    skip_config:

    ; --- PERMISOS: Dar control total a la carpeta de DATOS ---
    nsExec::ExecToLog 'icacls "$DataDir" /grant *S-1-5-32-545:(OI)(CI)F /T'
    nsExec::ExecToLog 'icacls "$DataDir" /grant *S-1-5-20:(OI)(CI)F /T'

    ; 2. Prerrequisitos (VC++ Redistributable)
    DetailPrint "Verificando librerias del sistema..."
    SetOutPath "$INSTDIR"
    File "../../dist/windows/vc_redist.x64.exe"
    ExecWait '"$INSTDIR\vc_redist.x64.exe" /install /quiet /norestart'
    Delete "$INSTDIR\vc_redist.x64.exe"

    ; 3. Configurar PostgreSQL
    DetailPrint "Extrayendo motor de base de datos..."
    SetOutPath "$INSTDIR"
    File "../../dist/windows/pgsql.zip"
    nsisunz::Unzip "$INSTDIR\pgsql.zip" "$INSTDIR"
    Delete "$INSTDIR\pgsql.zip"

    ; --- ROBUSTEZ: Añadir bin a PATH temporalmente ---
    ReadEnvStr $0 PATH
    StrCpy $0 "$INSTDIR\pgsql\bin;$0"
    System::Call 'Kernel32::SetEnvironmentVariable(t "PATH", t r0)'

    ; VERIFICAR SI LA BASE DE DATOS YA ESTA INICIALIZADA
    IfFileExists "$DataDir\data\PG_VERSION" db_already_exists

    DetailPrint "Inicializando base de datos local por primera vez..."
    FileOpen $0 "$INSTDIR\pw.txt" w
    FileWrite $0 "password"
    FileClose $0

    nsExec::ExecToLog '"$INSTDIR\pgsql\bin\initdb.exe" -D "$DataDir\data" -U postgres --pwfile="$INSTDIR\pw.txt" -E UTF8 -A scram-sha-256'
    Delete "$INSTDIR\pw.txt"

    ; CONFIGURACION DE RED
    FileOpen $0 "$DataDir\data\postgresql.conf" a
    FileSeek $0 0 END
    FileWrite $0 "$\r$\nlisten_addresses = '*'"
    FileClose $0

    FileOpen $0 "$DataDir\data\pg_hba.conf" a
    FileSeek $0 0 END
    FileWrite $0 "$\r$\nhost    all             all             127.0.0.1/32            scram-sha-256"
    FileWrite $0 "$\r$\nhost    all             all             ::1/128                 scram-sha-256"
    FileClose $0

    db_already_exists:

    DetailPrint "Registrando/Actualizando servicio VerithDB..."

    ; Verificar si el binario existe
    IfFileExists "$INSTDIR\pgsql\bin\pg_ctl.exe" +2 0
        DetailPrint "ERROR CRÍTICO: No se encontró pg_ctl.exe en $INSTDIR\pgsql\bin"
    nsExec::ExecToLog 'sc delete VerithDB'
    nsExec::ExecToStack '"$INSTDIR\pgsql\bin\pg_ctl.exe" register -N "VerithDB" -D "$DataDir\data" -S auto -w'
    
    Pop $0 ; Exit Code
    Pop $1 ; Output
    DetailPrint "Salida de pg_ctl: $1"
    DetailPrint "Código de salida: $0"

    DetailPrint "Iniciando servicio..."
    nsExec::ExecToLog 'net start VerithDB'
    
    ; Intentar crear la base de datos siempre (si ya existe, fallará silenciosamente, lo cual está bien)
    DetailPrint "Asegurando base de datos verithdb..."
    Sleep 5000 
    System::Call 'Kernel32::SetEnvironmentVariable(t "PGPASSWORD", t "password")'
    nsExec::Exec '"$INSTDIR\pgsql\bin\createdb.exe" -U postgres verithdb'
    
    ; Si la base de datos es nueva (cluster nuevo), configuración adicional si fuera necesaria
    IfFileExists "$DataDir\data\PG_VERSION" skip_db_creation
        ; (Espacio reservado para lógica futura de cluster nuevo, por ahora vacío ya que createdb está arriba)
    skip_db_creation:

    ; 5. Registro y Accesos Directos
    WriteUninstaller "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Verith" "Install_Dir" "$INSTDIR"
    WriteRegStr HKLM "Software\Verith" "Data_Dir" "$DataDir"

    CreateDirectory "$SMPROGRAMS\Verith"
    CreateShortcut "$SMPROGRAMS\Verith\Verith.lnk" "$INSTDIR\Verith.exe"
    CreateShortcut "$SMPROGRAMS\Verith\Desinstalar Verith.lnk" "$INSTDIR\Uninstall.exe"
    CreateShortcut "$SMPROGRAMS\Desinstalar Verith.lnk" "$INSTDIR\Uninstall.exe"
    CreateShortcut "$DESKTOP\Verith.lnk" "$INSTDIR\Verith.exe"

    ; Panel de Control
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Verith" "DisplayName" "Verith - Gestion Financiera"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Verith" "UninstallString" "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Verith" "DisplayIcon" "$INSTDIR\Verith.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Verith" "Publisher" "Verith Software"
SectionEnd

; --- Uninstaller Steps ---
Section "Uninstall"
    ReadRegStr $DataDir HKLM "Software\Verith" "Data_Dir"

    DetailPrint "Deteniendo y eliminando servicio VerithDB..."
    nsExec::ExecToLog 'sc stop VerithDB'
    Sleep 2000 
    nsExec::ExecToLog 'sc delete VerithDB'

    Delete "$DESKTOP\Verith.lnk"
    Delete "$SMPROGRAMS\Desinstalar Verith.lnk"
    RMDir /r "$SMPROGRAMS\Verith"
    RMDir /r "$INSTDIR"

    MessageBox MB_YESNO "¿Desea eliminar también la base de datos y la configuración? (No podrá recuperar sus datos)" IDNO skip_data
        RMDir /r "$DataDir"
    skip_data:

    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Verith"
    DeleteRegKey HKLM "Software\Verith"
SectionEnd
