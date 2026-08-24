/**
  @Author: dongxx
  @Since: 2025/8/18 17:09
  @Desc: //TODO
**/

package asynqpackage

// 任务队列类型, 注册消费者的时候, 需要指定消费队列及优先级, 如果没有指定则只会消费默认队列的任务
// 如果一个任务投递到了队列a, 且消费队列a的服务没有注册相应的处理handler, 则任务会失败: handler not found for task "xxx"
var schoolConvert *SchoolConvertQueue
var manageService *ManageServiceQueue
var defaultQueue *DefaultQueue

func Init() {
	if schoolConvert == nil {
		schoolConvert = &SchoolConvertQueue{
			Queue:            "school_convert",   //队列名称
			RecordDocxExport: "recordDocxExport", // 回放相关导出, 导出点评, 导出纪要, 导出转写 都会使用这个任务类型, 通过任务内部参数区分
			RecordDownload:   "recordDownload",   // 回放下载
			RecordTaskClear:  "recordTaskClear",  // 回放下载任务过期清理
			RemoveFile:       "removeFile",       // 删除文件
		}
	}
	if manageService == nil {
		manageService = &ManageServiceQueue{
			Queue:         "manage_service", //队列名称
			TeacherImport: "teacherImport",  // 批量导入老师
			StudentImport: "studentImport",  // 批量导入学生
			ClassImport:   "classImport",    // 批量导入教室
		}
	}
	if defaultQueue == nil {
		defaultQueue = &DefaultQueue{
			Queue: "default",
		}
	}
}

func GetSchoolConvertQueue() *SchoolConvertQueue {
	if schoolConvert == nil {
		Init()
	}
	return schoolConvert
}
func GetManageServiceQueue() *ManageServiceQueue {
	if manageService == nil {
		Init()
	}
	return manageService
}
func GetDefaultQueueQueue() *DefaultQueue {
	if defaultQueue == nil {
		Init()
	}
	return defaultQueue
}

type SchoolConvertQueue struct {
	Queue            string
	RecordDocxExport string
	RecordDownload   string
	RecordTaskClear  string
	RemoveFile       string
}

type ManageServiceQueue struct {
	Queue         string
	TeacherImport string
	StudentImport string
	ClassImport   string
}

type DefaultQueue struct {
	Queue string
}
