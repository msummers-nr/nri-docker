@setlocal enableextensions
@cd /d "%~dp0"
echo New Relic OHI Installer

@echo off
goto do_Install

:do_Install
    net session >nul 2>&1
    if %errorLevel% == 0 (
        echo Success: Administrative permissions confirmed.
        copy windows\docker-config.yml "C:\Program Files\New Relic\newrelic-infra\custom-integrations\"
        copy windows\docker-ohi.exe "C:\Program Files\New Relic\newrelic-infra\custom-integrations\"
        copy windows\docker-ohi-definition.yml "C:\Program Files\New Relic\newrelic-infra\custom-integrations\"
        copy windows\docker-ohi-config.yml "C:\Program Files\New Relic\newrelic-infra\integrations.d\"

        net stop newrelic-infra
        net start newrelic-infra
    ) else (
        echo Failure: Administrative permissions required!
    )

timeout 5