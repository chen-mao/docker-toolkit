#ifndef __XDX_SMI_API_H__
#define __XDX_SMI_API_H__

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdbool.h>

#define MAX_DEVICES 8

#define PARAMETER_PAGE_OFFSET  0x58000
#define FW_GLOABL_STRUCT_OFFSET  0x59000
#define CMCU_VARS_OFFSET       0x6FF000   //7MB-4KB

/**
 * Buffer size guaranteed to be large enough for \ref xdxmlDeviceGetName
 */
#define XDXML_DEVICE_NAME_BUFFER_SIZE  64

/**
 * Buffer size guaranteed to be large enough for pci bus id
 */
#define XDXML_DEVICE_PCI_BUS_ID_BUFFER_SIZE 32


/**
 *
 *GPU wcj
*/
#define XDXML_XDXGPU_INFO_UTIL_INFO                         0x02
#define XDXML_RGXFWIF_DM_MAX                                    (9U)
#define XDXML_RGX_NUM_OS_SUPPORTED                          1

/**
 *
 * PCI
*/
#define XDXML_DEVICE_PCI_BUS_ID_BUFFER_SIZE 32

/**
 *
 *gpu mem wcj
*/
#define RET_SIZE 1000000
#define MAX_PATH_LEN 100
#define PROCESS_NAME_LEN 30

/** 
 * Return values for XDXML API calls. 
 */
typedef enum XDXML_return_enum 
{
    // cppcheck-suppress *
    XDXML_ERROR = -1,
    XDXML_SUCCESS = 0,                   //!< The operation was successful
    XDXML_ERROR_UNINITIALIZED = 1,       //!< XDXML was not first initialized with XDXMLInit()
    XDXML_ERROR_INVALID_ARGUMENT = 2,    //!< A supplied argument is invalid
    XDXML_ERROR_NOT_SUPPORTED = 3,       //!< The requested operation is not available on target device
    XDXML_ERROR_NO_PERMISSION = 4,       //!< The current user does not have permission for operation
    XDXML_ERROR_ALREADY_INITIALIZED = 5, //!< Deprecated: Multiple initializations are now allowed through ref counting
    XDXML_ERROR_NOT_FOUND = 6,           //!< A query to find an object was unsuccessful
    XDXML_ERROR_INSUFFICIENT_SIZE = 7,   //!< An input argument is not large enough
    XDXML_ERROR_INSUFFICIENT_POWER = 8,  //!< A device's external power cables are not properly attached
    XDXML_ERROR_DRIVER_NOT_LOADED = 9,   //!< xdx driver is not loaded
    XDXML_ERROR_TIMEOUT = 10,            //!< User provided timeout passed
    XDXML_ERROR_IRQ_ISSUE = 11,          //!< xdx Kernel detected an interrupt issue with a GPU
    XDXML_ERROR_LIBRARY_NOT_FOUND = 12,  //!< XDXML Shared Library couldn't be found or loaded
    XDXML_ERROR_FUNCTION_NOT_FOUND = 13, //!< Local version of XDXML doesn't implement this function
    XDXML_ERROR_CORRUPTED_INFOROM = 14,  //!< infoROM is corrupted
    XDXML_ERROR_GPU_IS_LOST = 15,        //!< The GPU has fallen off the bus or has otherwise become inaccessible
    XDXML_ERROR_RESET_REQUIRED = 16,     //!< The GPU requires a reset before it can be used again
    XDXML_ERROR_OPERATING_SYSTEM = 17,   //!< The GPU control device has been blocked by the operating system/cgroups
    XDXML_ERROR_LIB_RM_VERSION_MISMATCH = 18,   //!< RM detects a driver/library version mismatch
    XDXML_ERROR_IN_USE = 19,             //!< An operation cannot be performed because the GPU is currently in use
    XDXML_ERROR_MEMORY = 20,             //!< Insufficient memory
    XDXML_ERROR_NO_DATA = 21,            //!<No data
    XDXML_ERROR_VGPU_ECC_NOT_SUPPORTED = 22,    //!< The requested vgpu operation is not available on target device, becasue ECC is enabled
    XDXML_ERROR_UNKNOWN = 999            //!< An internal driver error occurred
} xdxml_return_t;

typedef struct node_path_st{
    char card[128];
    char hwmon[128];
    char name[256];
    char dbdf[256];
    char current_link_speed[256];
    char current_link_width[256];
    char max_link_speed[256];
    char max_link_width[256];
} node_path_st;

struct xdx_device_st{
    struct node_path_st node_path;
    unsigned int index;
    char architecture[64];
    char pci_dbdf[64];
    char uuid[10];
    uint64_t domain;
    uint64_t bus;
    uint64_t device;
    uint64_t func;
    uint64_t bdfid;

    unsigned int curr_link_speed;
    unsigned int curr_link_width;
    unsigned int mx_link_speed;
    unsigned int mx_link_width;
    unsigned int curr_link_gen;
    unsigned int mx_link_gen;
	struct pci_device *dev;
	uint64_t device_id;
	// static gpu info
    int xdx_gddr_phys_freq;                       //!< GPU显存的频率
    int xdx_gfx_freq;                             //!< GPU的频率，单位为HZ
    int xdx_acpu_freq;                            //!< GPU HOST的频率
    int xdx_cmcu_freq;                            //!< GPU支持的GFX频率
    int xdx_ndf_freq;                             //!< GPU NDF的频率
    char name[XDXML_DEVICE_NAME_BUFFER_SIZE];       //!< GPU产品名称
    char vbios_version[64];                         //!< GPU vbios版本号
    char fw_ver[64];                                //!< 显卡固件版本 
    unsigned long long fb_total;// 显存大小    
};

typedef struct {
    struct xdx_device_st* handle;
} xdx_device_t;

/**
 * PCI information about a GPU device.
 */
typedef struct xdx_pci_info_st
{
    uint64_t domain;             //!< The PCI domain on which the device's bus resides, 0 to 0xffffffff
    uint64_t bus;                //!< The bus on which the device resides, 0 to 0xff
    uint64_t device;             //!< The device's id on the bus, 0 to 31
    uint64_t bdfid;              //bus device function id
    uint64_t func;
    char pci_dbdf[64];
    char bus_id[XDXML_DEVICE_PCI_BUS_ID_BUFFER_SIZE];
} xdx_pci_info_t;

/**
 * 该函数位于api调用流程的前端，调用其它api前应首先调用该函数初始化so库。该函数会确认驱动节点状态，若初始化失败，
 * 则无法通过get_handle获取device结构体。
 *
 * @brief                               初始化so库，在调用其它api前应首先调用该api
 * 
 * @param NULL
 *
 * @return
 *         - \ref XDXML_SUCCESS         successful completion          
 */
xdxml_return_t xdxml_init();

/**
 * 调用该函数后会清空所有device结构体内容。之前通过如 \ref xdxml_device_get_handle_by_index()
 * 所申请的device里的内容同样会被清零。在调用shutdown后若需调用其他api，则需要再次进行初始化
 * \ref xdxml_init() 并重新申请device结构体。
 *
 * @param NULL
 *
 * @return
 *         - \ref XDXML_SUCCESS         successful completion          
 */
xdxml_return_t xdxml_shutdown();

 /**
 * 获取当前系统中存在的所有XdxGPU的数量
 * @brief Retrieves the number of compute devices in the system. A compute device is a single GPU.
 * 
 * @param deviceCount                           Reference in which to return the number of accessible devices
 * 
 * @return 
 *         - \ref XDXML_SUCCESS                 if \a deviceCount has been set
 *         - \ref XDXML_ERROR_INVALID_ARGUMENT  if \a deviceCount is NULL
 * 
 */
xdxml_return_t xdxml_device_get_count(unsigned int *device_count);

/**
 * 获取GPU产品名称，此处产品名称用架构指代，如：PANGU
 *
 * @param device                        A structure used to save device information
 * @param device_name                   A poiter used to save gpu product name    
 * 
 * @return
 *         - \ref XDXML_SUCCESS         successful completion          
 */
xdxml_return_t xdxml_device_get_product_name(xdx_device_t device,char *device_name);

/***************************************************************************************************/
/** @defgroup xdxml_device_queries Device Queries
 * This chapter describes that queries that XDXML can perform against each device.
 * In each case the device is identified with an xdxml_device_t handle. This handle is obtained by  
 * calling one of \ref xdxml_device_get_handle_by_index(), \ref xdxml_device_get_handle_by_serial(),
 * \ref xdxml_device_get_handle_by_pci_bus_id(). or \ref xdxml_device_get_handle_by_uuid(). 
 *  @{
 */
/***************************************************************************************************/

/**
* 通过index参数获取对应的GPU device结构体，并用于后续其他api的传参调用。device结构体中包含对应gpu所拥有的
* 如文件节点路径、pcie信息、序列号等必要信息。index索引的范围 >= 0且< device_num即识别到的所有gpu个数，device_num请使用
* xdxml_device_get_count()函数获取。
* 
* Note: 使用本函数以及其它函数前，请首先调用 \ref xdxml_init() 函数。
* 
* @brief                                       通过index参数获取device结构体变量
*  
* @param index                                 The index of the target GPU, >= 0 and < \a accessibleDevices
* @param device                                Reference in which to return the device handle
* 
* @return 
*         - \ref XDXML_SUCCESS                 if \a deviceCount has been set
*         - \ref XDXML_ERROR_INVALID_ARGUMENT  if \a deviceCount is NULL
*/
xdxml_return_t xdxml_device_get_handle_by_index(unsigned int index, xdx_device_t* device);

/**
 * Get gpu uuid value
 *
 * @param device                        A structure used to save device information    
 * 
 * @return
 *         - \ref XDXML_SUCCESS         successful completion          
 */
xdxml_return_t xdxml_device_get_uuid(xdx_device_t device);

/**
 * 获得GPU的次设备号即minor number，在XDX GPU中，minor number即用index编号来指代
 *
 *
 * @param device                        A structure used to save device information
 * @param minor_num                     A poiter used to save gpu minor_num    
 * 
 * @return
 *         - \ref XDXML_SUCCESS         successful completion          
 */
xdxml_return_t xdxml_device_get_minor_number(xdx_device_t device,int *minor_num);

// New version add
/**
 * 获取对应GPU的PCIe相关数据。该函数会将数据填充至传入的结构体中，并返回是否处理成功。
 * 
 * 详见 \ref  xdxml_pci_info_t 查看可以获取的pcie数据。
 * 
 * @brief                                       get the dbdf of PCI，domain:bus:device.function
 *
 * @param device                                The identifier of the target device
 * @param pci                                   Reference in which to return the PCI info
 * 
 * @return 
 *         - \ref XDXML_SUCCESS                 if \a name has been set
 *         - \ref XDXML_ERROR_INVALID_ARGUMENT  if \a device is invalid, or \a name is NULL
 */
xdxml_return_t xdxml_device_get_pci_info(xdx_device_t device, xdx_pci_info_t *pci);

#ifdef __cplusplus
}
#endif //__cplusplus

#endif //__XDX_SMI_API_H__