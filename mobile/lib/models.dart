class FileItem {
  final String name;
  final String path;
  final bool isDir;
  final int size;
  final int modTime;
  final int capabilities;

  FileItem({
    required this.name,
    required this.path,
    required this.isDir,
    required this.size,
    required this.modTime,
    required this.capabilities,
  });

  factory FileItem.fromJson(Map<String, dynamic> json) {
    return FileItem(
      name: json['name'] as String,
      path: json['path'] as String,
      isDir: json['is_dir'] as bool,
      size: json['size'] as int,
      modTime: json['mod_time'] as int,
      capabilities: json['capabilities'] as int,
    );
  }

  static const int capBrowse = 1;
  static const int capStream = 2;
  static const int capRender = 4;
  static const int capEdit = 8;

  bool get canBrowse => (capabilities & capBrowse) != 0;
  bool get canStream => (capabilities & capStream) != 0;
  bool get canRender => (capabilities & capRender) != 0;
  bool get canEdit => (capabilities & capEdit) != 0;
}

class ListResponse {
  final List<FileItem> items;
  final String effectivePath;

  ListResponse({
    required this.items,
    required this.effectivePath,
  });

  factory ListResponse.fromJson(Map<String, dynamic> json) {
    return ListResponse(
      items: (json['items'] as List<dynamic>)
          .map((e) => FileItem.fromJson(e as Map<String, dynamic>))
          .toList(),
      effectivePath: json['effective_path'] as String,
    );
  }
}
