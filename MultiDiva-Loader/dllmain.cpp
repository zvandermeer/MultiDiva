// dllmain.cpp : Defines the entry point for the DLL application.
#define WIN32_LEAN_AND_MEAN
#include <windows.h>

// Standard
#include <string>
#include <iostream>
#include <fstream> 

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

#define STB_IMAGE_IMPLEMENTATION
#include "deps/stb_image.h"

enum NoteGrade {
	Cool = 0,
	Good = 1,
	Safe = 2,
	Cool_Wrong = 3,
	Good_Wrong = 4
};

struct NoteData {
	bool connectedPlayer;
	int32_t fullScore;
	int32_t slicedScore[7];
	int32_t combo;
	NoteGrade grade;
};

struct MultiDivaInit_return {
	bool* connectedToServer;
	bool* connectedToRoom;
	int* score;
};

struct Image {
	ID3D11ShaderResourceView* texture;
	int width;
	int height;
};

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
bool show_ingame_gui = false;
bool show_fullscreen_menu = false;
bool show_endgame_menu = false;
bool f10Pressed = false;
bool f9Pressed = false;
bool f8Pressed = false;
bool publicRoom = true;
bool* connectedToServer;
bool* connectedToRoom;
float speed = 0.5;

MultiDivaInit_return myInitReturn;

bool init = false;
HWND window = NULL;
ID3D11Device* p_device = NULL;
ID3D11DeviceContext* p_context = NULL;
ID3D11RenderTargetView* mainRenderTargetView = NULL;

static char serverAddress[128] = "";
static char serverPort[5] = "9988";
static char roomName[128] = "";

static char pushNotification[256] = "";
static char serverStatus[256] = "";
static char serverStatusTooltip[256] = "";
static char roomStatus[256] = "";
static char serverVersion[5] = "";

NoteData playerScoresForUI[10];

// Mod Library
HMODULE m_Library;

// Mod Types
typedef MultiDivaInit_return(__cdecl* _OnInit)(char* pushNotification, char* serverStatus, char* serverStatusTooltip, char* roomStatus, char* serverVersion, NoteData[]);
typedef void(__cdecl* _OnDispose)();
typedef void(__cdecl* _OnSongUpdate)(int songId, bool isPractice);
typedef void(__cdecl* _MainLoop)();
typedef void(__cdecl* _OnScoreTrigger)();
typedef void(__cdecl* _ConnectToServer)(const char* serverAddress, const char* serverPort);
typedef void(__cdecl* _LeaveServer)();
typedef void(__cdecl* _JoinRoom)(const char* roomName);
typedef void(__cdecl* _CreateRoom)(const char* roomName, bool publicRoom);


// Mod Functions
_OnInit p_OnInit;
_OnDispose p_OnDispose;
_OnSongUpdate p_OnSongUpdate;
_MainLoop p_MainLoop;
_OnScoreTrigger p_OnScoreTrigger;
_ConnectToServer p_ConnectToServer;
_LeaveServer p_LeaveServer;
_JoinRoom p_JoinRoom;
_CreateRoom p_CreateRoom;

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

/*
 * Hooks
 */
HOOK(void, __fastcall, _SongStart, sigSongStart, int songId)
{
	if (m_Library)
	{
		// Playing
		p_OnSongUpdate(songId, false);

		show_ingame_gui = true;
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

		show_ingame_gui = false;
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
		p_ConnectToServer = (_ConnectToServer)GetProcAddress(m_Library, "ConnectToServer");
		p_LeaveServer = (_LeaveServer)GetProcAddress(m_Library, "LeaveServer");
		p_JoinRoom = (_JoinRoom)GetProcAddress(m_Library, "JoinRoom");
		p_CreateRoom = (_CreateRoom)GetProcAddress(m_Library, "CreateRoom");

		// Install Hooks
		INSTALL_HOOK(_SongStart);
		INSTALL_HOOK(_SongEnd);
		INSTALL_HOOK(_SongPracticeStart);
		INSTALL_HOOK(_PrintResult);

		NoteData myGamer;
		myGamer.combo = 12;
		myGamer.fullScore = 7258571;
		myGamer.grade = Good;
		myGamer.slicedScore[0] = 7;
		myGamer.slicedScore[1] = 2;
		myGamer.slicedScore[2] = 5;
		myGamer.slicedScore[3] = 8;
		myGamer.slicedScore[4] = 5;
		myGamer.slicedScore[5] = 7;
		myGamer.slicedScore[6] = 1;

		NoteData myGamer2;
		myGamer2.combo = 42;
		myGamer2.fullScore = 370;
		myGamer2.grade = Safe;
		myGamer2.slicedScore[0] = 3;
		myGamer2.slicedScore[1] = 7;
		myGamer2.slicedScore[2] = 0;

		playerScoresForUI[0] = myGamer;
		playerScoresForUI[1] = myGamer2;

		// Mod Entry Point
		myInitReturn = p_OnInit(pushNotification, serverStatus, serverStatusTooltip, roomStatus, serverVersion, playerScoresForUI);

		connectedToServer = myInitReturn.connectedToServer;
		connectedToRoom = myInitReturn.connectedToRoom;
	}
}

extern "C" __declspec(dllexport) void OnFrame() {
	if (p_MainLoop) {
		p_MainLoop();
	}
}

// Simple helper function to load an image into a DX11 texture with common settings
bool LoadTextureFromFile(const char* filename, ID3D11ShaderResourceView** out_srv, int* out_width, int* out_height)
{
	// Load from disk into a raw RGBA buffer
	int image_width = 0;
	int image_height = 0;
	unsigned char* image_data = stbi_load(filename, &image_width, &image_height, NULL, 4);
	if (image_data == NULL)
		return false;

	// Create texture
	D3D11_TEXTURE2D_DESC desc;
	ZeroMemory(&desc, sizeof(desc));
	desc.Width = image_width;
	desc.Height = image_height;
	desc.MipLevels = 1;
	desc.ArraySize = 1;
	desc.Format = DXGI_FORMAT_R8G8B8A8_UNORM;
	desc.SampleDesc.Count = 1;
	desc.Usage = D3D11_USAGE_DEFAULT;
	desc.BindFlags = D3D11_BIND_SHADER_RESOURCE;
	desc.CPUAccessFlags = 0;

	ID3D11Texture2D* pTexture = NULL;
	D3D11_SUBRESOURCE_DATA subResource;
	subResource.pSysMem = image_data;
	subResource.SysMemPitch = desc.Width * 4;
	subResource.SysMemSlicePitch = 0;
	p_device->CreateTexture2D(&desc, &subResource, &pTexture);

	// Create texture view
	D3D11_SHADER_RESOURCE_VIEW_DESC srvDesc;
	ZeroMemory(&srvDesc, sizeof(srvDesc));
	srvDesc.Format = DXGI_FORMAT_R8G8B8A8_UNORM;
	srvDesc.ViewDimension = D3D11_SRV_DIMENSION_TEXTURE2D;
	srvDesc.Texture2D.MipLevels = desc.MipLevels;
	srvDesc.Texture2D.MostDetailedMip = 0;
	p_device->CreateShaderResourceView(pTexture, &srvDesc, out_srv);
	pTexture->Release();

	*out_width = image_width;
	*out_height = image_height;
	stbi_image_free(image_data);

	return true;
}

int my_image_width = 0;
int my_image_height = 0;
Image coolTexture;

Image myNumbers[10];

ImFont* exoFontXL;
ImFont* exoFontLarge;
ImFont* exoFontMedium;
ImFont* exoFontSmall;
ImFont* defaultFont;

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

			exoFontXL = io.Fonts->AddFontFromFileTTF("mods\\MultiDiva\\assets\\font\\Exo2-Regular.ttf", 40);
			exoFontLarge = io.Fonts->AddFontFromFileTTF("mods\\MultiDiva\\assets\\font\\Exo2-Regular.ttf", 30);
			exoFontMedium = io.Fonts->AddFontFromFileTTF("mods\\MultiDiva\\assets\\font\\Exo2-Regular.ttf", 20);
			exoFontSmall = io.Fonts->AddFontFromFileTTF("mods\\MultiDiva\\assets\\font\\Exo2-Regular.ttf", 16);
			defaultFont = io.Fonts->AddFontDefault();

			bool epicBool = LoadTextureFromFile("mods\\MultiDiva\\assets\\img\\cool.png", &coolTexture.texture, &coolTexture.width, &coolTexture.height);
			IM_ASSERT(epicBool);

			for (int i = 0; i < 10; i++) {
				std::string s = "mods\\MultiDiva\\assets\\img\\num\\" + std::to_string(i) + ".png";
				bool newEpicBool = LoadTextureFromFile(s.c_str(), &myNumbers[i].texture, &myNumbers[i].width, &myNumbers[i].height);
				IM_ASSERT(newEpicBool);
			}


		}
		else
			return p_present(p_swap_chain, sync_interval, flags);
	}

	ImGui_ImplDX11_NewFrame();
	ImGui_ImplWin32_NewFrame();

	ImGui::NewFrame();

	ImGui::PushFont(defaultFont);

	ImGui::ShowDemoWindow(); // check demo cpp

	//if (show_menu) {
	//	ImGui::Begin("Test Env", &show_menu);
	//	ImGui::SetWindowSize(ImVec2(200, 200), ImGuiCond_Always);
	//	ImGui::Text("Options:");
	//	ImGui::Checkbox("Canvas", &show_canvas);
	//	ImGui::Checkbox("Show MultiDiva", &show_my_menu);
	//	ImGui::SliderFloat("Speed", &speed, 0.01, 1);
	//	ImGui::End();
	//}
	//if (show_my_menu) {
	//	ImGui::Begin("MultiDiva", &show_my_menu);
	//	ImGui::SetWindowSize(ImVec2(200, 200), ImGuiCond_FirstUseEver);
	//	ImGui::Text("Options:");
	//	ImGui::InputText("input text", str0, IM_ARRAYSIZE(str0));
	//	ImGui::InputText("input text2", str1, IM_ARRAYSIZE(str1));
	//	ImGui::End();
	//}

	if (show_endgame_menu) {
		static ImGuiWindowFlags flags = ImGuiWindowFlags_NoDecoration | ImGuiWindowFlags_NoMove | ImGuiWindowFlags_NoResize | ImGuiWindowFlags_NoSavedSettings;
		const ImGuiViewport* viewport = ImGui::GetMainViewport();
		ImGui::SetNextWindowPos(viewport->Pos);
		ImGui::SetNextWindowSize(viewport->Size);
		ImGui::PushStyleColor(ImGuiCol_WindowBg, ImVec4(0.08f, 0.08f, 0.08f, 0.95f));
		ImGui::PushFont(exoFontXL);
		ImGui::Begin("Endgame", &show_fullscreen_menu, flags);

		ImGui::NewLine();
		ImGui::Indent();
		ImGui::Indent();
		ImGui::Text("Results");

		ImGui::PushFont(exoFontLarge);
		ImGui::NewLine();
		if (ImGui::BeginTable("table1", 3))
		{
			for (int row = 0; row < 4; row++)
			{
				ImGui::TableNextRow();
				for (int column = 0; column < 3; column++)
				{
					ImGui::TableSetColumnIndex(column);
					ImGui::Text("Row %d Column %d", row, column);
				}
			}
			ImGui::EndTable();
		}
		ImGui::End();
	}

	if (show_fullscreen_menu) {
		static ImGuiWindowFlags flags = ImGuiWindowFlags_NoDecoration | ImGuiWindowFlags_NoMove | ImGuiWindowFlags_NoResize | ImGuiWindowFlags_NoSavedSettings;
		const ImGuiViewport* viewport = ImGui::GetMainViewport();
		ImGui::SetNextWindowPos(viewport->Pos);
		ImGui::SetNextWindowSize(viewport->Size);
		ImGui::PushStyleColor(ImGuiCol_WindowBg, ImVec4(0.08f, 0.08f, 0.08f, 0.95f));
		ImGui::Begin("Test Fullscreen", &show_fullscreen_menu, flags); 
		ImGui::NewLine();
		ImGui::Indent();
		ImGui::Text("MultiDiva v");
		ImGui::SameLine(0,0);
		ImGui::Text(serverVersion);
		if (ImGui::CollapsingHeader("Server")) {
			ImGui::Text("Server address: ");
			ImGui::InputText("##serverAddressInput", serverAddress, IM_ARRAYSIZE(serverAddress));
			ImGui::Text("Server port: ");
			ImGui::InputText("##serverPortInput", serverPort, IM_ARRAYSIZE(serverPort));
			if (!*connectedToServer) {
				if (ImGui::Button("Connect")) {
					if (m_Library) {
						p_ConnectToServer(serverAddress, serverPort);
					}
				}
			}
			else {
				if (ImGui::Button("Disconnect")) {
					if (m_Library) {
						p_LeaveServer();
					}
				}
			}
			ImGui::SameLine();
			ImGui::Text(serverStatus);
			if (ImGui::IsItemHovered() && strcmp(serverStatusTooltip, "") != 0) {
				ImGui::SetTooltip(serverStatusTooltip);
			}
		}
		if (!*connectedToServer) {
			ImGui::BeginDisabled();
		}
		if (ImGui::CollapsingHeader("Room")) {
			if (*connectedToRoom) {
				ImGui::BeginDisabled();
			}
			ImGui::Text("Room name: ");
			ImGui::InputText("##roomNameInput", roomName, IM_ARRAYSIZE(roomName));
			ImGui::Text("Public room? ");
			ImGui::SameLine();
			ImGui::Checkbox("##publicRoomCheckbox", &publicRoom);
			if (*connectedToRoom) {
				ImGui::EndDisabled();
			}
			if (!*connectedToRoom) {
				if (ImGui::Button("Join")) {
					if (m_Library) {
						p_JoinRoom(roomName);
					}
				}
				ImGui::SameLine();
				if (ImGui::Button("Create")) {
					if (m_Library) {
						p_CreateRoom(roomName, true);
					}
				}
			}
			else {
				if (ImGui::Button("Leave room")) {

				}
			}
			ImGui::SameLine();
			ImGui::Text(roomStatus);
		}
		if (!*connectedToServer) {
			ImGui::EndDisabled();
		}
		if (ImGui::CollapsingHeader("Funny Pictures")) {
			ImGui::Text("pointer = %p", coolTexture);
			ImGui::Text("size = %d x %d", my_image_width, my_image_height);
			ImGui::Image((void*)coolTexture.texture, ImVec2(coolTexture.width, coolTexture.height));
			for (int i = 0; i < 10; i++) {
				ImGui::Image((void*)myNumbers[i].texture, ImVec2(32, 32));
				ImGui::SameLine();
			}
			ImGui::Text("Show in game gui?");
			ImGui::Checkbox("##inGameGUI", &show_ingame_gui);
			ImGui::Checkbox("Canvas", &show_canvas);
		}
		ImGui::End();
	}

	if (show_ingame_gui) {
		for (int i = 0; i < 10; i++) {
			if (playerScoresForUI[i].connectedPlayer == true) {
				static ImGuiWindowFlags flags = ImGuiWindowFlags_NoDecoration | ImGuiWindowFlags_NoMove | ImGuiWindowFlags_NoResize;
				const ImGuiViewport* viewport = ImGui::GetMainViewport();

				float offset = (viewport->Size.y / 12) + (viewport->Size.y / 48);

				ImGui::SetNextWindowSize(ImVec2(viewport->Size.x / 6.4, viewport->Size.y / 12));
				ImGui::SetNextWindowPos(ImVec2(viewport->Size.x - (viewport->Size.x / 5.5), viewport->Pos.y + (viewport->Size.y / 14.4) + (offset * i)));
				ImGui::PushStyleColor(ImGuiCol_WindowBg, ImVec4(0.08f, 0.08f, 0.08f, 0.20f));
				char integer_string[32];
				sprintf_s(integer_string, "%d", i);
				char other_string[64] = "In game GUI";
				strcat_s(other_string, integer_string);
				ImGui::Begin(other_string, &show_ingame_gui, flags);
				ImGui::PushFont(exoFontMedium);
				ImGui::Text("UsernameHere");
				ImGui::PopFont;

				for (int j = 0; j < 7; j++) {
					ImGui::Image((void*)myNumbers[playerScoresForUI[i].slicedScore[j]].texture, ImVec2(viewport->Size.x / 64, viewport->Size.x / 64));
					if (viewport->Size.x == 1280) {
						ImGui::SameLine(0, 0.1);
					}
					else {
						ImGui::SameLine(0, 0);
					}
				}
				ImGui::End();
			}
		}
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

		if (GetAsyncKeyState(VK_F10) && !f10Pressed) {
			show_fullscreen_menu = !show_fullscreen_menu;
			f10Pressed = true;
		}
		if (!GetAsyncKeyState(VK_F10)) {
			f10Pressed = false;
		}

		if (GetAsyncKeyState(VK_F9) && !f9Pressed) {
			show_ingame_gui = !show_ingame_gui;
			f9Pressed = true;
		}
		if (!GetAsyncKeyState(VK_F9)) {
			f9Pressed = false;
		}

		if (GetAsyncKeyState(VK_F8) && !f8Pressed) {
			show_endgame_menu = !show_endgame_menu;
			f8Pressed = true;
		}
		if (!GetAsyncKeyState(VK_F8)) {
			f8Pressed = false;
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

BOOL APIENTRY DllMain(HMODULE hModule,
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
