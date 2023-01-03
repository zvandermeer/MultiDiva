// dllmain.cpp : Defines the entry point for the DLL application.
#include "pch.h"
#include <SigScan.h>
#include "Diva.h"

// 1.02 (RocketRacer)
const uint64_t DivaScoreBaseAddress = 0x00000001412EF568;
const uint64_t DivaScoreCompletionRateAddress = 0x00000001412EF634;
const uint64_t DivaScoreWorstCounterAddress = 0x00000001416E2D40; // For whatever reason the "worst" counter is stored separately from the rest of the hit counters
const uint64_t DivaScoreGradeAddress = 0x00000001416E2D00;
const uint64_t DivaCurrentPVTitleAddress = 0x00000001412EF228;
const uint64_t DivaCurrentPVIdAddress = 0x00000001412C2340;
const uint64_t DivaCurrentPVDifficultyAddress = 0x00000001423157AC;

// Mod Library
HMODULE m_Library;

// Mod Types
typedef void(__cdecl* _OnInit)();
typedef void(__cdecl* _OnDispose)();
typedef void(__cdecl* _OnSongUpdate)(int songId, bool isPractice);
typedef void(__cdecl* _MainLoop)();
typedef void(__cdecl* _OnScoreTrigger)(int, int);


// Mod Functions
_OnInit p_OnInit;
_OnDispose p_OnDispose;
_OnSongUpdate p_OnSongUpdate;
_MainLoop p_MainLoop;
_OnScoreTrigger p_OnScoreTrigger;

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


// 1.02: 0x14043B2D0 (Braasileiro)
void* sigSongStart = sigScan(
    "\x8B\xD1\xE9\xA9\xE8\xFF\xFF\xCC\xCC\xCC\xCC\xCC\xCC\xCC\xCC\xCC\xE9",
    "xxxxxxx?????????x"
);

// 1.02: 0x1401E7A60 (Braasileiro)
void* sigSongPracticeStart = sigScan(
    "\xE9\x00\x00\x00\x00\xA3\xF6\x42\xF3\xF8\x58\xFD\x35\x1D",
    "x????xxxxxxxxx"
);

// 1.02: 0x14043B000 (Braasileiro)
void* sigSongEnd = sigScan(
    "\x48\x89\x5C\x24\x08\x57\x48\x83\xEC\x20\x48\x8D\x0D\xCC\xCC\xCC\xCC\xE8\xCC\xCC\xCC\xCC\x48\x8B\x3D\xCC\xCC\xCC\xCC\x48\x8B\x1F\x48\x3B\xDF",
    "xxxxxxxxxxxxx????x????xxx????xxxxxx"
);

// 1.02 (RocketRacer)
void* DivaScoreTrigger = sigScan(
    "\x48\x89\x5C\x24\x00\x48\x89\x74\x24\x00\x48\x89\x7C\x24\x00\x55\x41\x54\x41\x55\x41\x56\x41\x57\x48\x8B\xEC\x48\x83\xEC\x60\x48\x8B\x05\x00\x00\x00\x00\x48\x33\xC4\x48\x89\x45\xF8\x48\x8B\xF9\x80\xB9\x00\x00\x00\x00\x00\x0F\x85\x00\x00\x00\x00",
    "xxxx?xxxx?xxxx?xxxxxxxxxxxxxxxxxxx????xxxxxxxxxxxx?????xx????"
);

/*
 * Hooks
 */
HOOK(void, __fastcall, _SongStart, sigSongStart, int songId)
{
    if (m_Library)
    {
        // Playing
        p_OnSongUpdate(songId, false);
    }

    original_SongStart(songId);
}

HOOK(__int64, __fastcall, _SongPracticeStart, sigSongPracticeStart, __int64 a1, __int64 a2)
{
    if (m_Library)
    {
        // Practicing
        p_OnSongUpdate(0, true);
    }

    return original_SongPracticeStart(a1, a2);
}

HOOK(__int64, __stdcall, _SongEnd, sigSongEnd)
{
    if (m_Library)
    {
        // In Menu
        p_OnSongUpdate(0, false);
    }

    return original_SongEnd();
}

HOOK(int, __fastcall, _PrintResult, DivaScoreTrigger, int a1) {
    int difficulty;
    int grade;

    std::string& DivaTitle = *(std::string*)DivaCurrentPVTitleAddress;
    DIVA_DIFFICULTY DivaDiff = *(_DIVA_DIFFICULTY*)DivaCurrentPVDifficultyAddress;
    DIVA_GRADE DivaGrade = *(_DIVA_GRADE*)DivaScoreGradeAddress;

    switch (DivaDiff)
    {
        case Normal:
            difficulty = 1;
            break;
        case Hard:
            difficulty = 2;
            break;
        case Extreme:
            difficulty = 3;
            break;
        case ExExtreme:
            difficulty = 4;
            break;
        case Easy:
            difficulty = 0;
            break;
        default:
            break;
    }
    switch (DivaGrade)
    {
        case Failed:
            grade = 0;
            break;
        case Cheap:
            grade = 1;
            break;
        case Standard:
            grade = 2;
            break;
        case Great:
            grade = 3;
            break;
        case Excellent:
            grade = 4;
            break;
        case Perfect:
            grade = 5;
            break;
        default:
            break;
    }

    if (m_Library) {
        p_OnScoreTrigger(difficulty, grade);
    }

    return original_PrintResult(a1);
}

/*
 * ModLoader
 */
extern "C" __declspec(dllexport) void Init()
{
    // Load Mod Library
    m_Library = LoadLibraryA("MultiDiva-Client.dll");

    if (m_Library)
    {
        // Mod Function Pointers
        p_OnInit = (_OnInit)GetProcAddress(m_Library, "MultiDivaInit");
        p_OnDispose = (_OnDispose)GetProcAddress(m_Library, "MultiDivaDispose");
        p_OnSongUpdate = (_OnSongUpdate)GetProcAddress(m_Library, "SongUpdate");
        p_MainLoop = (_MainLoop)GetProcAddress(m_Library, "MainLoop");
        p_OnScoreTrigger = (_OnScoreTrigger)GetProcAddress(m_Library, "OnScoreTrigger");

        // Install Hooks
        INSTALL_HOOK(_SongStart);
        INSTALL_HOOK(_SongEnd);
        INSTALL_HOOK(_SongPracticeStart);
        INSTALL_HOOK(_PrintResult);

        // Mod Entry Point
        p_OnInit();
    }
}

extern "C" __declspec(dllexport) void OnFrame() {
    if (p_MainLoop) {
        p_MainLoop();
    }
}
