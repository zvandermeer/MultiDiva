// Source: https://github.com/RocketRacer/MegaMixPlusShareDivaMod/blob/main/Diva.h

typedef enum _DIVA_DIFFICULTY : uint32_t {
    Easy = 0x0,
    Normal = 0x1,
    Hard = 0x2,
    Extreme = 0x3,
    ExExtreme = 0x4,
} DIVA_DIFFICULTY;
typedef enum _DIVA_GRADE : uint32_t {
    Failed = 0x0,
    Cheap = 0x1,
    Standard = 0x2,
    Great = 0x3,
    Excellent = 0x4,
    Perfect = 0x5
} DIVA_GRADE;
typedef struct _DIVA_PV_DIF {
    unsigned int Difficulty;
} DIVA_PV_DIF;
typedef struct _DIVA_PV_ID {
    unsigned int Id;
} DIVA_PV_ID;
typedef struct _DIVA_STAT {
    float CompletionRate;
} DIVA_STAT;
typedef struct _DIVA_SCORE {
    unsigned int TotalScore;
    unsigned int Unknown1;
    unsigned int Unknown2;
    unsigned int Unknown3;
    unsigned int Unknown4;
    unsigned int Unknown5;
    unsigned int Unknown6;
    unsigned int Unknown7;
    unsigned int Unknown8;
    unsigned int Combo;
    unsigned int preAdjustCool;
    unsigned int preAdjustFine;
    unsigned int preAdjustSafe;
    unsigned int preAdjustSad;
    unsigned int preAdjustWorst;
    unsigned int Cool;
    unsigned int Fine;
    unsigned int Safe;
    unsigned int Sad;
    unsigned int Worst;
} DIVA_SCORE;