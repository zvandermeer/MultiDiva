// dllmain.cpp : Defines the entry point for the DLL application.
#include "pch.h"

// Mod Library
HMODULE m_Library;

// Mod Types
typedef void(__cdecl* _OnInit)();
typedef void(__cdecl* _OnDispose)();
typedef void(__cdecl* _OnSongUpdate)(int songId, bool isPractice);
typedef void(__cdecl* _MainLoop)();
typedef void(__cdecl* _NoteHit)();

// Mod Functions
_OnInit p_OnInit;
_OnDispose p_OnDispose;
_OnSongUpdate p_OnSongUpdate;
_MainLoop p_MainLoop;
_NoteHit p_NoteHit;

BOOL APIENTRY DllMain( HMODULE hModule,
                       DWORD  ul_reason_for_call,
                       LPVOID lpReserved
                     )
{
    switch (ul_reason_for_call)
    {
    case DLL_PROCESS_ATTACH:
        break;
    case DLL_THREAD_ATTACH:
        break;
    case DLL_THREAD_DETACH:
        break;
    case DLL_PROCESS_DETACH:
        if (m_Library) {
            p_OnDispose();
        }
        break;
    }
    return TRUE;
}


/*
 * Signatures
 */

 // v1.02: 0x14040B270 (ActualMandM)
 //        0x14040B260 (Direct Signature) (Braasileiro)
SIG_SCAN
(
    sigSongData,
    0x14040B260,
    "\x48\x8D\x05\xCC\xCC\xCC\xCC\xC3\xCC\xCC\xCC\xCC\xCC\xCC\xCC\xCC\x48\x89\x5C\x24\x08\x48\x89\x6C\x24\x10\x48\x89\x74\x24\x18\x57\x48\x83\xEC\x20",
    "xxx????x????????xxxxxxxxxxxxxxxxxxxx"
)

// 1.02: 0x14043B2D0 (Braasileiro)
SIG_SCAN
(
    sigSongStart,
    0x14043B2D0,
    "\x8B\xD1\xE9\xA9\xE8\xFF\xFF\xCC\xCC\xCC\xCC\xCC\xCC\xCC\xCC\xCC\xE9",
    "xxxxxxx?????????x"
);

// 1.02: 0x1401E7A60 (Braasileiro)
SIG_SCAN
(
    sigSongPracticeStart,
    0x1401E7A60,
    "\xE9\x00\x00\x00\x00\xA3\xF6\x42\xF3\xF8\x58\xFD\x35\x1D",
    "x????xxxxxxxxx"
);

// 1.02: 0x14043B000 (Braasileiro)
SIG_SCAN
(
    sigSongEnd,
    0x14043B000,
    "\x48\x89\x5C\x24\x08\x57\x48\x83\xEC\x20\x48\x8D\x0D\xCC\xCC\xCC\xCC\xE8\xCC\xCC\xCC\xCC\x48\x8B\x3D\xCC\xCC\xCC\xCC\x48\x8B\x1F\x48\x3B\xDF",
    "xxxxxxxxxxxxx????x????xxx????xxxxxx"
);


/*
 * Hooks
 */
HOOK(void, __fastcall, _SongStart, sigSongStart(), int songId)
{
    if (m_Library)
    {
        // Playing
        p_OnSongUpdate(songId, false);
    }

    original_SongStart(songId);
}

HOOK(__int64, __fastcall, _SongPracticeStart, sigSongPracticeStart(), __int64 a1, __int64 a2)
{
    if (m_Library)
    {
        // Practicing
        p_OnSongUpdate(0, true);
    }

    return original_SongPracticeStart(a1, a2);
}

HOOK(__int64, __stdcall, _SongEnd, sigSongEnd())
{
    if (m_Library)
    {
        // In Menu
        p_OnSongUpdate(0, false);
    }

    return original_SongEnd();
}


/*
 * ModLoader
 */
extern "C" __declspec(dllexport) void Init()
{
    std::cout << "MultiDiva starting..." << std::endl;

    // Load Mod Library
    m_Library = LoadLibraryA("MultiDiva-Client.dll");

    if (m_Library)
    {
        // Mod Function Pointers
        p_OnInit = (_OnInit)GetProcAddress(m_Library, "MultiDivaInit");
        p_OnDispose = (_OnDispose)GetProcAddress(m_Library, "MultiDivaDispose");
        p_OnSongUpdate = (_OnSongUpdate)GetProcAddress(m_Library, "PrintSong");
        p_MainLoop = (_MainLoop)GetProcAddress(m_Library, "MainLoop");

        // Install Hooks
        INSTALL_HOOK(_SongStart);
        INSTALL_HOOK(_SongEnd);
        INSTALL_HOOK(_SongPracticeStart);

        // Current PID
        auto pid = GetCurrentProcessId();

        // 1.02: 0x14040B260
        auto addr = (uint8_t*)sigSongData();

        // 1.02: 0x1416E2B89
        auto pointer = (uintptr_t)(addr + ReadUnalignedU32(addr + 0x3));

        // Mod Entry Point
        p_OnInit();
    }
}

extern "C" __declspec(dllexport) void OnFrame() {
    if (p_MainLoop) {
        p_MainLoop();
    }
}
