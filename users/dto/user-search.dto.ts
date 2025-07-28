import { ApiProperty } from '@nestjs/swagger';
import { IsOptional, IsString, IsNumberString } from 'class-validator';

export class UserSearchQueryDto {
  @ApiProperty({
    example: 'John',
    description: 'The search term to query for users.',
    required: false,
  })
  @IsOptional()
  @IsString()
  q?: string;

  @ApiProperty({
    example: 1,
    description: 'Optional: Page number for pagination.',
    required: false,
  })
  @IsOptional()
  @IsNumberString()
  pageNumber?: string;

  @ApiProperty({
    example: 20,
    description: 'Optional: Number of records per page.',
    required: false,
  })
  @IsOptional()
  @IsNumberString()
  numberOfRecords?: string;
}
