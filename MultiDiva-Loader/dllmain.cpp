// dllmain.cpp : Defines the entry point for the DLL application.
#define WIN32_LEAN_AND_MEAN
#include <windows.h>

// Standard
#include <string>
#include <iostream>
#include <fstream> 
#include <dinput.h>

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

char* initStr(int len) {
	char* newStr = (char*)malloc(len);
	if (newStr) newStr[0] = '\0';
	return newStr;
}

enum NoteGrade {
	Cool = 0,
	Good = 1,
	Safe = 2,
	Cool_Wrong = 3,
	Good_Wrong = 4
};

struct UIPlayerScore {
	bool connectedPlayer;
	bool isThisPlayer;
	char* username;
	int32_t fullScore;
	int32_t slicedScore[7];
	int32_t combo;
	NoteGrade grade;

	UIPlayerScore() {
		username = initStr(1);
	}
};

struct InGameMenu {
	// Menu Status
	bool menuVisible;

	// Menu Data
	UIPlayerScore scores[10];
};

InGameMenu myInGameMenu;

struct EndgameMenu {
	bool menuVisible;
};

EndgameMenu myEndgameMenu;

struct ConnectionMenu {
	// Menu Status
	bool menuVisible;
	bool connectedToServer;
	bool connectedToRoom;

	// Display Labels
	char* pushNotification;
	char* serverStatus;
	char* serverStatusTooltip;
	char* roomStatus;
	char* serverVersion;

	// Input Fields
	char* serverAddress;
	char* serverPort;
	char* roomName;

	ConnectionMenu() {
		pushNotification = initStr(1);
		serverStatus = initStr(1);
		serverStatusTooltip = initStr(1);
		roomStatus = initStr(1);
		serverVersion = initStr(1);
		serverAddress = initStr(128);
		serverPort = initStr(5);
		roomName = initStr(128);
	}
};

ConnectionMenu myConnectionMenu;


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

bool show_canvas = false;
bool f10Pressed = false;
bool f9Pressed = false;
bool f8Pressed = false;
bool publicRoom = true;
float speed = 0.5;

bool init = false;
HWND window = NULL;
ID3D11Device* p_device = NULL;
ID3D11DeviceContext* p_context = NULL;
ID3D11RenderTargetView* mainRenderTargetView = NULL;

// Mod Library
HMODULE m_Library;

// Mod Types
typedef void(__cdecl* _OnInit)(ConnectionMenu*, InGameMenu*, EndgameMenu*);
typedef void(__cdecl* _OnDispose)();
typedef void(__cdecl* _OnSongUpdate)(int songId, bool isPractice);
typedef void(__cdecl* _MainLoop)();
typedef void(__cdecl* _OnScoreTrigger)();
typedef void(__cdecl* _ConnectToServer)(const char* serverAddress, const char* serverPort);
typedef void(__cdecl* _LeaveServer)();
typedef void(__cdecl* _JoinRoom)(const char* roomName);
typedef void(__cdecl* _CreateRoom)(const char* roomName, bool publicRoom);
//typedef void(_cdecl* _SongRunning)();


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
//_SongRunning p_SongRunning;

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

//// 1.03: 0x1511D321E
//void* GameRunningTrigger = sigScan(
//	"\x44\x8b\x81\x00\x00\x00\x00\x89\xd0",
//	"xxx????xx"
//);

/*
 * Hooks
 */
HOOK(void, __fastcall, _SongStart, sigSongStart, int songId)
{
	if (m_Library)
	{
		// Playing
		p_OnSongUpdate(songId, false);

		myInGameMenu.menuVisible = true;
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

		myInGameMenu.menuVisible = false;
	}

	return original_PrintResult(a1);
}

//uintptr_t p = 0x14CC0CB00;
//
//HOOK(void, __fastcall, _SongIsRunning, GameRunningTrigger, __int64 a1, int a2, __int64 a3, char a4) {
//	int gamePaused = *reinterpret_cast<int*>(p);
//
//	if (gamePaused == 0) {
//		INPUT ip;
//		ip.type = INPUT_KEYBOARD;
//		ip.ki.time = 0;
//		ip.ki.wVk = 0;
//		ip.ki.dwExtraInfo = 0;
//		ip.ki.dwFlags = KEYEVENTF_SCANCODE;
//		ip.ki.wScan = DIKEYBOARD_ESCAPE;
//		SendInput(1, &ip, sizeof(INPUT));
//		Sleep(50); // Sleep 50 milliseconds before key up
//		ip.ki.dwFlags = KEYEVENTF_KEYUP; // set the flag so the key goes up so it doesn't repeat keys
//		SendInput(1, &ip, sizeof(INPUT)); // Resend the input
//
//		if (m_Library) {
//			p_SongRunning();
//		}
//	}
//	
//	original_SongIsRunning(a1, a2, a3, a4);
//}

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
		//p_SongRunning = (_SongRunning)GetProcAddress(m_Library, "SongRunning");

		// Install Hooks
		INSTALL_HOOK(_SongStart);
		INSTALL_HOOK(_SongEnd);
		INSTALL_HOOK(_SongPracticeStart);
		INSTALL_HOOK(_PrintResult);
		//INSTALL_HOOK(_SongIsRunning);

		UIPlayerScore myGamer;
		myGamer.connectedPlayer = true;
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

		UIPlayerScore myGamer2;
		myGamer2.connectedPlayer = true;
		myGamer2.combo = 42;
		myGamer2.fullScore = 370;
		myGamer2.grade = Safe;
		myGamer2.slicedScore[0] = 0;
		myGamer2.slicedScore[1] = 0;
		myGamer2.slicedScore[2] = 0;
		myGamer2.slicedScore[3] = 0;
		myGamer2.slicedScore[4] = 3;
		myGamer2.slicedScore[5] = 7;
		myGamer2.slicedScore[6] = 0;

		myInGameMenu.scores[0] = myGamer;
		myInGameMenu.scores[1] = myGamer2;

		//strcpy_s(myConnectionMenu.serverAddress, "I love Kasane Teto so much");

		// Mod Entry Point
		p_OnInit(&myConnectionMenu, &myInGameMenu, &myEndgameMenu);
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

	if (myEndgameMenu.menuVisible) {
		static ImGuiWindowFlags flags = ImGuiWindowFlags_NoDecoration | ImGuiWindowFlags_NoMove | ImGuiWindowFlags_NoResize | ImGuiWindowFlags_NoSavedSettings;
		const ImGuiViewport* viewport = ImGui::GetMainViewport();
		ImGui::SetNextWindowPos(viewport->Pos);
		ImGui::SetNextWindowSize(viewport->Size);
		ImGui::PushStyleColor(ImGuiCol_WindowBg, ImVec4(0.08f, 0.08f, 0.08f, 0.95f));
		ImGui::PushFont(exoFontXL);
		ImGui::Begin("Endgame", &myEndgameMenu.menuVisible, flags);

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

	if (myConnectionMenu.menuVisible) {
		static ImGuiWindowFlags flags = ImGuiWindowFlags_NoDecoration | ImGuiWindowFlags_NoMove | ImGuiWindowFlags_NoResize | ImGuiWindowFlags_NoSavedSettings;
		const ImGuiViewport* viewport = ImGui::GetMainViewport();
		ImGui::SetNextWindowPos(viewport->Pos);
		ImGui::SetNextWindowSize(viewport->Size);
		ImGui::PushStyleColor(ImGuiCol_WindowBg, ImVec4(0.08f, 0.08f, 0.08f, 0.95f));
		ImGui::Begin("Test Fullscreen", &myConnectionMenu.menuVisible, flags);
		ImGui::NewLine();
		ImGui::Indent();
		ImGui::Text("MultiDiva v");
		ImGui::SameLine(0, 0);
		ImGui::Text(myConnectionMenu.serverVersion);
		if (ImGui::CollapsingHeader("Server")) {
			ImGui::Text("Server address: ");
			ImGui::InputText("##serverAddressInput", myConnectionMenu.serverAddress, IM_ARRAYSIZE(myConnectionMenu.serverAddress));
			ImGui::Text("Server port: ");
			ImGui::InputText("##serverPortInput", myConnectionMenu.serverPort, IM_ARRAYSIZE(myConnectionMenu.serverPort));
			if (!myConnectionMenu.connectedToServer) {
				if (ImGui::Button("Connect")) {
					if (m_Library) {
						p_ConnectToServer(myConnectionMenu.serverAddress, myConnectionMenu.serverPort);
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
			ImGui::Text(myConnectionMenu.serverStatus);
			if (ImGui::IsItemHovered() && strcmp(myConnectionMenu.serverStatusTooltip, "") != 0) {
				ImGui::SetTooltip(myConnectionMenu.serverStatusTooltip);
			}
		}
		if (!myConnectionMenu.connectedToServer) {
			ImGui::BeginDisabled();
		}
		if (ImGui::CollapsingHeader("Room")) {
			if (myConnectionMenu.connectedToRoom) {
				ImGui::BeginDisabled();
			}
			ImGui::Text("Room name: ");
			ImGui::InputText("##roomNameInput", myConnectionMenu.roomName, IM_ARRAYSIZE(myConnectionMenu.roomName));
			ImGui::Text("Public room? ");
			ImGui::SameLine();
			ImGui::Checkbox("##publicRoomCheckbox", &publicRoom);
			if (myConnectionMenu.connectedToRoom) {
				ImGui::EndDisabled();
			}
			if (!myConnectionMenu.connectedToRoom) {
				if (ImGui::Button("Join")) {
					if (m_Library) {
						p_JoinRoom(myConnectionMenu.roomName);
					}
				}
				ImGui::SameLine();
				if (ImGui::Button("Create")) {
					if (m_Library) {
						p_CreateRoom(myConnectionMenu.roomName, true);
					}
				}
			}
			else {
				if (ImGui::Button("Leave room")) {

				}
			}
			ImGui::SameLine();
			ImGui::Text(myConnectionMenu.roomStatus);
		}
		if (!myConnectionMenu.connectedToServer) {
			ImGui::EndDisabled();
		}
		if (ImGui::CollapsingHeader("Funny Pictures")) {
			ImGui::Text("pointer = %p", coolTexture);
			ImGui::Text("size = %d x %d", my_image_width, my_image_height);
			ImGui::Image((void*)coolTexture.texture, ImVec2(coolTexture.width, coolTexture.height));
			for (int i = 0; i < 10; i++) {
				if(i != 0) ImGui::SameLine();
				ImGui::Image((void*)myNumbers[i].texture, ImVec2(32, 32));
			}
			ImGui::Text("Show in game gui?");
			ImGui::SameLine();
			ImGui::Checkbox("##inGameGUI", &myInGameMenu.menuVisible);
			ImGui::Checkbox("Canvas", &show_canvas);
		}
		ImGui::End();
	}

	if (myInGameMenu.menuVisible) {
		for (int i = 0; i < 10; i++) {
			UIPlayerScore thisPlayer = myInGameMenu.scores[i];
			if (thisPlayer.connectedPlayer == true) {
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
				ImGui::Begin(other_string, &myInGameMenu.menuVisible, flags);
				ImGui::PushFont(exoFontMedium);
				ImGui::Text(thisPlayer.username);
				ImGui::PopFont;

				for (int j = 0; j < 7; j++) {
					ImGui::Image((void*)myNumbers[thisPlayer.slicedScore[j]].texture, ImVec2(viewport->Size.x / 64, viewport->Size.x / 64));
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
			myConnectionMenu.menuVisible = !myConnectionMenu.menuVisible;
			f10Pressed = true;
		}
		if (!GetAsyncKeyState(VK_F10)) {
			f10Pressed = false;
		}

		if (GetAsyncKeyState(VK_F9) && !f9Pressed) {
			myInGameMenu.menuVisible = !myInGameMenu.menuVisible;
			f9Pressed = true;
		}
		if (!GetAsyncKeyState(VK_F9)) {
			f9Pressed = false;
		}

		if (GetAsyncKeyState(VK_F8) && !f8Pressed) {
			myEndgameMenu.menuVisible = !myEndgameMenu.menuVisible;
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
