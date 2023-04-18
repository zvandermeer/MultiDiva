// dllmain.cpp : Defines the entry point for the DLL application.
#define WIN32_LEAN_AND_MEAN
#include <windows.h>

// Standard
#include <string>
#include <iostream>

// Deps
#include "deps/helpers.h"
#include "deps/detours/include/detours.h"
#include "deps/SigScan.h"

#include "deps/MinHook/MinHook.h"
#if _WIN64 
#pragma comment(lib, "deps/MinHook/libMinHook.x64.lib")
#else
#pragma comment(lib, "deps/MinHook/libMinHook.x86.lib")
#endif

#include "deps/ImGui/imgui.h"
#include "deps/ImGui/imgui_impl_win32.h"
#include "deps/ImGui/imgui_impl_dx11.h"
#include "deps/ImGui/imgui_internal.h" // for free drawing

#include <d3d11.h>
#pragma comment(lib, "d3d11.lib")

// Globals
HINSTANCE dll_handle;

typedef long(__stdcall* present)(IDXGISwapChain*, UINT, UINT);
present p_present;
present p_present_target;
bool get_present_pointer()
{
	DXGI_SWAP_CHAIN_DESC sd;
	ZeroMemory(&sd, sizeof(sd));
	sd.BufferCount = 2;
	sd.BufferDesc.Format = DXGI_FORMAT_R8G8B8A8_UNORM;
	sd.BufferUsage = DXGI_USAGE_RENDER_TARGET_OUTPUT;
	sd.OutputWindow = GetForegroundWindow();
	sd.SampleDesc.Count = 1;
	sd.Windowed = TRUE;
	sd.SwapEffect = DXGI_SWAP_EFFECT_DISCARD;

	IDXGISwapChain* swap_chain;
	ID3D11Device* device;

	const D3D_FEATURE_LEVEL feature_levels[] = { D3D_FEATURE_LEVEL_11_0, D3D_FEATURE_LEVEL_10_0, };
	if (D3D11CreateDeviceAndSwapChain(
		NULL,
		D3D_DRIVER_TYPE_HARDWARE,
		NULL,
		0,
		feature_levels,
		2,
		D3D11_SDK_VERSION,
		&sd,
		&swap_chain,
		&device,
		nullptr,
		nullptr) == S_OK)
	{
		void** p_vtable = *reinterpret_cast<void***>(swap_chain);
		swap_chain->Release();
		device->Release();
		//context->Release();
		p_present_target = (present)p_vtable[8];
		return true;
	}
	return false;
}

WNDPROC oWndProc;
// Win32 message handler your application need to call.
// - You should COPY the line below into your .cpp code to forward declare the function and then you can call it.
// - Call from your application's message handler. Keep calling your message handler unless this function returns TRUE.
// Forward declare message handler from imgui_impl_win32.cpp
extern LRESULT ImGui_ImplWin32_WndProcHandler(HWND hWnd, UINT msg, WPARAM wParam, LPARAM lParam);
LRESULT __stdcall WndProc(const HWND hWnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
	ImGuiIO& io = ImGui::GetIO();
	ImGui_ImplWin32_WndProcHandler(hWnd, uMsg, wParam, lParam);
	if (io.WantCaptureMouse && (uMsg == WM_LBUTTONDOWN || uMsg == WM_LBUTTONUP || uMsg == WM_RBUTTONDOWN || uMsg == WM_RBUTTONUP || uMsg == WM_MBUTTONDOWN || uMsg == WM_MBUTTONUP || uMsg == WM_MOUSEWHEEL || uMsg == WM_MOUSEMOVE))
	{
		return TRUE;
	}

	return CallWindowProc(oWndProc, hWnd, uMsg, wParam, lParam);
}


bool show_menu = true;
bool show_my_menu = true;
bool show_canvas = false;
float speed = 0.5;

bool init = false;
HWND window = NULL;
ID3D11Device* p_device = NULL;
ID3D11DeviceContext* p_context = NULL;
ID3D11RenderTargetView* mainRenderTargetView = NULL;

static char str0[128] = "Hello, world!";
static char str1[128] = "Hello, world!";

static long __stdcall detour_present(IDXGISwapChain* p_swap_chain, UINT sync_interval, UINT flags) {
	if (!init) {
		if (SUCCEEDED(p_swap_chain->GetDevice(__uuidof(ID3D11Device), (void**)&p_device)))
		{
			p_device->GetImmediateContext(&p_context);

			// Get HWND to the current window of the target/game
			DXGI_SWAP_CHAIN_DESC sd;
			p_swap_chain->GetDesc(&sd);
			window = sd.OutputWindow;

			// Location in memory where imgui is rendered to
			ID3D11Texture2D* pBackBuffer;
			p_swap_chain->GetBuffer(0, __uuidof(ID3D11Texture2D), (LPVOID*)&pBackBuffer);
			// create a render target pointing to the back buffer
			p_device->CreateRenderTargetView(pBackBuffer, NULL, &mainRenderTargetView);
			// This does not destroy the back buffer! It only releases the pBackBuffer object which we only needed to create the RTV.
			pBackBuffer->Release();
			oWndProc = (WNDPROC)SetWindowLongPtr(window, GWLP_WNDPROC, (LONG_PTR)WndProc);

			// Init ImGui 
			ImGui::CreateContext();
			ImGuiIO& io = ImGui::GetIO();
			//io.ConfigFlags = ImGuiConfigFlags_NoMouseCursorChange;
			ImGui_ImplWin32_Init(window);
			ImGui_ImplDX11_Init(p_device, p_context);
			init = true;
		}
		else
			return p_present(p_swap_chain, sync_interval, flags);
	}
	ImGui_ImplDX11_NewFrame();
	ImGui_ImplWin32_NewFrame();

	ImGui::NewFrame();

	ImGui::ShowDemoWindow(); // check demo cpp

	if (show_menu) {
		ImGui::Begin("Test Env", &show_menu);
		ImGui::SetWindowSize(ImVec2(200, 200), ImGuiCond_Always);
		ImGui::Text("Options:");
		ImGui::Checkbox("Canvas", &show_canvas);
		ImGui::Checkbox("Show MultiDiva", &show_my_menu);
		ImGui::SliderFloat("Speed", &speed, 0.01, 1);
		ImGui::End();
	}
	if (show_my_menu) {
		ImGui::Begin("MultiDiva", &show_my_menu);
		ImGui::SetWindowSize(ImVec2(200, 200), ImGuiCond_Always);
		ImGui::Text("Options:");
		ImGui::InputText("input text", str0, IM_ARRAYSIZE(str0));
		ImGui::InputText("input text2", str1, IM_ARRAYSIZE(str1));
		ImGui::End();
	}

	if (show_canvas) {
		ImGuiIO& io = ImGui::GetIO();

		ImGui::PushStyleVar(ImGuiStyleVar_WindowBorderSize, 0.0f);
		ImGui::PushStyleVar(ImGuiStyleVar_WindowPadding, { 0.0f, 0.0f });
		ImGui::PushStyleColor(ImGuiCol_WindowBg, { 0.0f, 0.0f, 0.0f, 0.0f });
		ImGui::Begin("XXX", nullptr, ImGuiWindowFlags_NoTitleBar | ImGuiWindowFlags_NoInputs);

		ImGui::SetWindowPos(ImVec2(0, 0), ImGuiCond_Always);
		ImGui::SetWindowSize(ImVec2(io.DisplaySize.x, io.DisplaySize.y), ImGuiCond_Always);

		ImGuiWindow* window = ImGui::GetCurrentWindow();
		ImDrawList* draw_list = window->DrawList;
		ImVec2 p0 = ImVec2(50, 25);
		ImVec2 p1 = ImVec2(200, 250);
		draw_list->AddRectFilled(p0, p1, IM_COL32(50, 50, 50, 255));
		draw_list->AddRect(p0, p1, IM_COL32(255, 255, 255, 255));

		ImVec2 midpoint = ImVec2(500, 500);
		draw_list->AddCircle(midpoint, 30, ImColor(51, 255, 0), 0, 20);

		window->DrawList->PushClipRectFullScreen();
		ImGui::End();
		ImGui::PopStyleColor();
		ImGui::PopStyleVar(2);

	}

	ImGui::EndFrame();

	// Prepare the data for rendering so we can call GetDrawData()
	ImGui::Render();

	p_context->OMSetRenderTargets(1, &mainRenderTargetView, NULL);
	// The real rendering
	ImGui_ImplDX11_RenderDrawData(ImGui::GetDrawData());

	return p_present(p_swap_chain, sync_interval, flags);
}

DWORD __stdcall EjectThread(LPVOID lpParameter) {
	Sleep(100);
	FreeLibraryAndExitThread(dll_handle, 0);
	Sleep(100);
	return 0;
}


//"main" loop
int WINAPI main()
{

	if (!get_present_pointer())
	{
		return 1;
	}

	MH_STATUS status = MH_Initialize();
	if (status != MH_OK)
	{
		return 1;
	}

	if (MH_CreateHook(reinterpret_cast<void**>(p_present_target), &detour_present, reinterpret_cast<void**>(&p_present)) != MH_OK) {
		return 1;
	}

	if (MH_EnableHook(p_present_target) != MH_OK) {
		return 1;
	}

	while (true) {
		Sleep(50);

		if (GetAsyncKeyState(VK_NUMPAD0) & 1) {

		}

		if (GetAsyncKeyState(VK_NUMPAD1)) {
			break;
		}
	}

	//Cleanup
	if (MH_DisableHook(MH_ALL_HOOKS) != MH_OK) {
		return 1;
	}
	if (MH_Uninitialize() != MH_OK) {
		return 1;
	}

	ImGui_ImplDX11_Shutdown();
	ImGui_ImplWin32_Shutdown();
	ImGui::DestroyContext();

	if (mainRenderTargetView) { mainRenderTargetView->Release(); mainRenderTargetView = NULL; }
	if (p_context) { p_context->Release(); p_context = NULL; }
	if (p_device) { p_device->Release(); p_device = NULL; }
	SetWindowLongPtr(window, GWLP_WNDPROC, (LONG_PTR)(oWndProc)); // "unhook"

	CreateThread(0, 0, EjectThread, 0, 0, 0);

	return 0;
}

// Mod Library
HMODULE m_Library;

// Mod Types
typedef void(__cdecl* _OnInit)();
typedef void(__cdecl* _OnDispose)();
typedef void(__cdecl* _OnSongUpdate)(int songId, bool isPractice);
typedef void(__cdecl* _MainLoop)();
typedef void(__cdecl* _OnScoreTrigger)();
typedef int(__cdecl* _TestFunc)();


// Mod Functions
_OnInit p_OnInit;
_OnDispose p_OnDispose;
_OnSongUpdate p_OnSongUpdate;
_MainLoop p_MainLoop;
_OnScoreTrigger p_OnScoreTrigger;
_TestFunc p_TestFunc;
// _OnNoteHit p_OnNoteHit;

BOOL APIENTRY DllMain( HMODULE hModule,
                       DWORD  ul_reason_for_call,
                       LPVOID lpReserved
                     )
{
    switch (ul_reason_for_call)
    {
    case DLL_PROCESS_ATTACH:
		dll_handle = hModule;
		CreateThread(NULL, 0, (LPTHREAD_START_ROUTINE)main, NULL, 0, NULL);
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
// 1.03: 0x14043B310 (Braasileiro)
void* sigSongStart = sigScan(
    "\x8B\xD1\xE9\xA9\xE8\xFF\xFF\xCC\xCC\xCC\xCC\xCC\xCC\xCC\xCC\xCC\xE9",
    "xxxxxxx?????????x"
);

// 1.02: 0x1401E7A60 (Braasileiro)
// 1.03: 0x1401E7A70 (Braasileiro)
void* sigSongPracticeStart = sigScan(
    "\xE9\x00\x00\x00\x00\x58\x3C\xB4",
    "x????xxx"
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

// 1.02: 0x14024319F
// void* sigNoteHit = sigScan(
//     "\x44\x0f\xb6\x65\x00\x44\x88\x64\x24",
//     "xxxx?xxxx"
// );

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

    if (m_Library) {
        p_OnScoreTrigger();
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
        p_TestFunc = (_TestFunc)GetProcAddress(m_Library, "TestFunc");
        // p_OnNoteHit = (_OnNoteHit)GetProcAddress(m_Library, "OnNoteHit");

		//get_present_pointer();

        // Install Hooks
        INSTALL_HOOK(_SongStart);
        INSTALL_HOOK(_SongEnd);
        INSTALL_HOOK(_SongPracticeStart);
        INSTALL_HOOK(_PrintResult);
        // INSTALL_HOOK(_NoteHit);

        // Mod Entry Point
        p_OnInit();

        /*int foo = p_TestFunc();*/
    }
}

extern "C" __declspec(dllexport) void OnFrame() {
    if (p_MainLoop) {
        p_MainLoop();
    }
}
