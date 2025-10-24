import * as Minio from 'minio';
import { Readable } from 'stream';
import PDFDocument from 'pdfkit';
import { Template1Data } from './document.types';
import { IMinioConfig, MinioConfigModel } from '../../database/models';

export class MinioService {
  private client: Minio.Client;
  private bucketName: string;
  private config: IMinioConfig;

  /**
   * Constructor - Tạo MinIO service với config
   * @param config - MinIO configuration từ MongoDB
   */
  constructor(config: IMinioConfig) {
    if (!config.enabled) {
      throw new Error('MinIO configuration is disabled');
    }

    this.config = config;

    // Initialize MinIO client
    this.client = new Minio.Client({
      endPoint: config.endPoint,
      port: config.port,
      useSSL: config.useSSL,
      accessKey: config.accessKey,
      secretKey: config.secretKey,
    });

    this.bucketName = config.bucketName;

    console.log(`MinIO initialized: ${config.endPoint}:${config.port}`);

    // Ensure bucket exists and update connection status
    this.ensureBucket()
      .then(async () => {
        // Update connection status to MongoDB
        await this.updateConnectionStatus(true);
      })
      .catch(async (err) => {
        console.error('Error ensuring bucket exists:', err);
        // Update connection status to MongoDB
        await this.updateConnectionStatus(false, err.message);
      });
  }

  private async ensureBucket(): Promise<void> {
    try {
      const exists = await this.client.bucketExists(this.bucketName);
      if (!exists) {
        const region = this.config?.region || 'us-east-1';
        await this.client.makeBucket(this.bucketName, region);
        console.log(`Bucket ${this.bucketName} created successfully`);
      }
    } catch (error) {
      console.error('Error ensuring bucket exists:', error);
      throw error;
    }
  }

  /**
   * Update connection status trong MongoDB
   */
  private async updateConnectionStatus(connected: boolean, error?: string): Promise<void> {
    try {
      await MinioConfigModel.findByIdAndUpdate(this.config._id, {
        connectionStatus: {
          connected,
          lastCheck: new Date(),
          error: error || undefined,
        },
      });
      console.log(`MinIO connection status updated: ${connected ? 'Connected' : 'Disconnected'}`);
    } catch (err) {
      console.error('Error updating MinIO connection status:', err);
    }
  }

  /**
   * Upload file từ Buffer vào MinIO
   * @param buffer - Buffer của file cần upload
   * @param filename - Tên file trong MinIO
   * @param contentType - Content type của file (default: application/octet-stream)
   * @returns Object name (path) trong MinIO
   */
  public async uploadBuffer(
    buffer: Buffer,
    filename: string,
    contentType: string = 'application/octet-stream'
  ): Promise<string> {
    try {
      const objectName = `${Date.now()}-${filename}`;
      const stream = Readable.from(buffer);

      const metaData = {
        'Content-Type': contentType,
      };

      await this.client.putObject(
        this.bucketName,
        objectName,
        stream,
        buffer.length,
        metaData
      );

      console.log(`File uploaded successfully: ${objectName}`);
      return objectName;
    } catch (error) {
      console.error('Error uploading buffer:', error);
      throw error;
    }
  }

  /**
   * Lấy file từ MinIO dưới dạng Buffer
   * @param objectName - Tên object trong MinIO
   * @returns Buffer của file
   */
  public async getFile(objectName: string): Promise<Buffer> {
    try {
      const dataStream = await this.client.getObject(this.bucketName, objectName);
      const chunks: Buffer[] = [];

      return new Promise((resolve, reject) => {
        dataStream.on('data', (chunk) => chunks.push(chunk));
        dataStream.on('end', () => resolve(Buffer.concat(chunks)));
        dataStream.on('error', reject);
      });
    } catch (error) {
      console.error('Error getting file:', error);
      throw error;
    }
  }

  /**
   * Lấy stream của file từ MinIO (tốt cho file lớn)
   * @param objectName - Tên object trong MinIO
   * @returns Readable stream
   */
  public async getFileStream(objectName: string): Promise<Readable> {
    try {
      return await this.client.getObject(this.bucketName, objectName);
    } catch (error) {
      console.error('Error getting file stream:', error);
      throw error;
    }
  }

  /**
   * Xóa file từ MinIO
   * @param objectName - Tên object cần xóa
   */
  public async deleteFile(objectName: string): Promise<void> {
    try {
      await this.client.removeObject(this.bucketName, objectName);
      console.log(`File deleted successfully: ${objectName}`);
    } catch (error) {
      console.error('Error deleting file:', error);
      throw error;
    }
  }

  /**
   * Xóa nhiều files cùng lúc
   * @param objectNames - Mảng tên objects cần xóa
   */
  public async deleteFiles(objectNames: string[]): Promise<void> {
    try {
      await this.client.removeObjects(this.bucketName, objectNames);
      console.log(`Files deleted successfully: ${objectNames.length} files`);
    } catch (error) {
      console.error('Error deleting files:', error);
      throw error;
    }
  }

  /**
   * Lấy presigned URL để download file (URL có thời hạn)
   * @param objectName - Tên object
   * @param expiry - Thời gian hết hạn (seconds), default 7 days
   * @returns URL để download
   */
  public async getPresignedUrl(
    objectName: string,
    expiry: number = 7 * 24 * 60 * 60
  ): Promise<string> {
    try {
      return await this.client.presignedGetObject(
        this.bucketName,
        objectName,
        expiry
      );
    } catch (error) {
      console.error('Error getting presigned URL:', error);
      throw error;
    }
  }

  /**
   * List tất cả files trong bucket
   * @param prefix - Filter theo prefix (optional)
   * @returns Mảng thông tin files
   */
  public async listFiles(prefix?: string): Promise<Minio.BucketItem[]> {
    try {
      const stream = this.client.listObjects(this.bucketName, prefix, true);
      const files: any[] = [];

      return new Promise((resolve, reject) => {
        stream.on('data', (obj) => files.push(obj));
        stream.on('end', () => resolve(files));
        stream.on('error', reject);
      });
    } catch (error) {
      console.error('Error listing files:', error);
      throw error;
    }
  }

  /**
   * Lấy thông tin chi tiết của file
   * @param objectName - Tên object
   * @returns Thông tin file (size, etag, lastModified, etc.)
   */
  public async getFileInfo(objectName: string): Promise<Minio.BucketItemStat> {
    try {
      return await this.client.statObject(this.bucketName, objectName);
    } catch (error) {
      console.error('Error getting file info:', error);
      throw error;
    }
  }

  /**
   * Check xem file có tồn tại không
   * @param objectName - Tên object
   * @returns true nếu tồn tại, false nếu không
   */
  public async fileExists(objectName: string): Promise<boolean> {
    try {
      await this.client.statObject(this.bucketName, objectName);
      return true;
    } catch (error: any) {
      if (error.code === 'NotFound') {
        return false;
      }
      throw error;
    }
  }

  /**
   * Tạo PDF từ Template 1 data và trả về Buffer
   * @param data - Data cho template 1
   * @returns Promise<Buffer> - PDF buffer
   */
  public async generateTemplate1PDF(data: Template1Data): Promise<Buffer> {
    return new Promise((resolve, reject) => {
      try {
        const doc = new PDFDocument({
          size: 'A4',
          margins: {
            top: 72,
            bottom: 72,
            left: 72,
            right: 72,
          },
        });

        const chunks: Buffer[] = [];

        // Collect PDF chunks
        doc.on('data', (chunk) => chunks.push(chunk));
        doc.on('end', () => resolve(Buffer.concat(chunks)));
        doc.on('error', reject);

        // Generate PDF content
        this.renderTemplate1(doc, data);

        // Finalize PDF
        doc.end();
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Render nội dung Template 1
   */
  private renderTemplate1(doc: PDFKit.PDFDocument, data: Template1Data): void {
    // Register Liberation Serif fonts - Times New Roman style (hỗ trợ tiếng Việt)
    const fontPath = '/usr/share/fonts/truetype/liberation';
    doc.registerFont('Regular', `${fontPath}/LiberationSerif-Regular.ttf`);
    doc.registerFont('Bold', `${fontPath}/LiberationSerif-Bold.ttf`);
    doc.registerFont('Italic', `${fontPath}/LiberationSerif-Italic.ttf`);
    doc.registerFont('BoldItalic', `${fontPath}/LiberationSerif-BoldItalic.ttf`);

    // Header - gốc trái, IN ĐẬM, cỡ 12
    doc.moveDown(1);
    doc
      .fontSize(12)
      .font('Bold')
      .text('TRƯỜNG ĐẠI HỌC BÁCH KHOA – ĐHQG TPHCM', { align: 'left' })
      .text('KHOA KHOA HỌC VÀ KỸ THUẬT MÁY TÍNH', { align: 'left' })
      .moveDown(1);

    // Title - căn giữa, IN ĐẬM, cỡ 12
    doc
      .fontSize(12)
      .font('Bold')
      .text('THÔNG TIN ĐỀ TÀI', { align: 'center' })
      .text('GIAI ĐOẠN 1 (GĐ1): ĐỀ CƯƠNG LUẬN VĂN/ ĐỒ ÁN CHUYÊN NGÀNH/', { align: 'center' })
      .text('ĐỒ ÁN MÔN HỌC KỸ THUẬT MÁY TÍNH', { align: 'center' })
      .text(`HỌC KỲ ${data.semester} NĂM HỌC ${data.academicYear}`, { align: 'center' })
      .moveDown(1);

    // Tên đề tài - gốc trái, IN ĐẬM, cỡ 12
    doc
      .fontSize(12)
      .font('Bold')
      .text('Tên đề tài:', { continued: false })
      .moveDown(0.3)
      .text(`- Tiếng Việt: ${data.thesisTitle.vietnamese}`)
      .text(`- Tiếng Anh: ${data.thesisTitle.english}`)
      .moveDown(1);

    // Thông tin công ty - IN ĐẬM, cỡ 12
    doc
      .fontSize(12)
      .font('Bold')
      .text(`Công ty/ Doanh nghiệp hợp tác: ${data.company?.name || 'None'}`)
      .text(`Địa chỉ: ${data.company?.address || '……………………………'}`)
      .text(`Website link: ${data.company?.websiteLink || '……………………'}`)
      .text(`Người đại diện giao tiếp với Khoa: ${data.company?.representativeName || '……………………'}`)
      .font('Italic')
      .text('(tối thiểu phải có thông tin họ tên, email công vụ/ cá nhân).')
      .moveDown(1);

    // Thông tin CBHD - cỡ 12
    data.teachers.forEach((teacher, index) => {
      const teacherNum = index + 1;
      doc
        .fontSize(12)
        .font('Regular')
        .text(`CBHD${teacherNum}: ${teacher.name}`, { continued: false })
        .text(`Email${teacherNum}: `, { continued: true, underline: false })
        .fillColor('blue')
        .text(teacher.email, { link: `mailto:${teacher.email}`, underline: true })
        .fillColor('black');

      if (index === 0) {
        doc
          .fontSize(12)
          .font('BoldItalic')
          .text('(Chuẩn CBHD1/ chính: ngạch giảng viên, đạt chuẩn giảng dạy lý thuyết.)');
      }
    });
    doc.moveDown(0.5);

    // Ngành - cỡ 12
    doc.fontSize(12).font('Regular').text('Ngành:');

    const majorsList = [
      { label: 'Khoa học máy tính', value: 'Khoa học máy tính' },
      { label: 'Kỹ thuật máy tính', value: 'Kỹ thuật máy tính' },
      { label: 'Liên ngành CS-CE', value: 'Liên ngành CS-CE' }
    ];

    const currentY1 = doc.y;
    const startX = doc.x;
    let currentX = startX;

    majorsList.forEach((m) => {
      const isChecked = m.value === data.major;

      // Draw checkbox
      doc.rect(currentX, currentY1, 10, 10).stroke();
      if (isChecked) {
        // Draw X for checked
        doc.moveTo(currentX + 2, currentY1 + 2)
           .lineTo(currentX + 8, currentY1 + 8)
           .moveTo(currentX + 8, currentY1 + 2)
           .lineTo(currentX + 2, currentY1 + 8)
           .stroke();
      }

      // Draw label
      doc.fontSize(12).font('Regular');
      doc.text(m.label, currentX + 15, currentY1, { continued: false, lineBreak: false });
      currentX += doc.widthOfString(m.label) + 60; // Move to next checkbox position
    });

    // Move cursor down after checkboxes
    doc.y = currentY1 + 15;
    doc.x = startX;

    // Chương trình đào tạo - cỡ 12 (xuống dòng)
    doc.fontSize(12).font('Regular').text('Chương trình đào tạo:');

    const programsList = [
      { label: 'Tiếng Việt (CQ/CN/B2/SN/VLVH/TX)', value: 'Tiếng Việt' },
      { label: 'Tiếng Anh (CC/CT/QT)', value: 'Tiếng Anh' }
    ];

    const currentY2 = doc.y;
    let currentX2 = startX;

    programsList.forEach((p) => {
      const isChecked = p.value === data.programLanguage;

      // Draw checkbox
      doc.rect(currentX2, currentY2, 10, 10).stroke();
      if (isChecked) {
        // Draw X for checked
        doc.moveTo(currentX2 + 2, currentY2 + 2)
           .lineTo(currentX2 + 8, currentY2 + 8)
           .moveTo(currentX2 + 8, currentY2 + 2)
           .lineTo(currentX2 + 2, currentY2 + 8)
           .stroke();
      }

      // Draw label
      doc.fontSize(12).font('Regular');
      doc.text(p.label, currentX2 + 15, currentY2, { continued: false, lineBreak: false });
      currentX2 += doc.widthOfString(p.label) + 60; // Move to next checkbox position
    });

    // Move cursor down after checkboxes
    doc.y = currentY2 + 15;
    doc.x = startX;
    doc.moveDown(0.5);

    // Số lượng sinh viên - cỡ 12
    doc
      .fontSize(12)
      .font('Regular')
      .text(`Số lượng sinh viên thực hiện: ${data.students.length}.`);

    // Danh sách sinh viên - cỡ 12, với major từ data
    data.students.forEach((student) => {
      doc
        .fontSize(12)
        .font('Regular')
        .text(`${student.name} - ${student.studentId} - ${student.program}`);
    });
    doc.moveDown(1);

    // Mô tả - người dùng nhập gì thì hiển thị y vậy
    if (data.description) {
      doc
        .fontSize(12)
        .font('Regular')
        .text(data.description, {
          align: 'left',
          lineGap: 2,
        });
    }
  }
}
