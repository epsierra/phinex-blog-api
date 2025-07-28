import {
  Controller,
  Get,
  Post,
  Put,
  Delete,
  Body,
  Param,
  Query,
  HttpCode,
  HttpStatus,
  UseGuards,
  Patch,
} from '@nestjs/common';
import { UsersService } from './users.service';
import {
  CreateUserDto,
  UserUpdateDto,
  CreateUserSettingDto,
  UpdateUserSettingDto,
  UpdateUserStatusDto,
  SubscribeToPlanDto,
} from './dto/user.dto';
import {
  ApiTags,
  ApiOperation,
  ApiResponse,
  ApiBearerAuth,
  ApiParam,
  ApiBody,
  ApiQuery,
} from '@nestjs/swagger';
import { CurrentUser } from '../common/decorators/current-user.decorator';
import { ICurrentUser } from 'src/common/constants';
import { AuthenticatedGuard } from 'src/common/guards/authenticated.guard';
import { AdminGuard } from 'src/common/guards/admin.guard';

@ApiTags('Users')
@Controller('users')
@ApiBearerAuth()
export class UsersController {
  constructor(private readonly usersService: UsersService) {}

  @Get()
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Retrieve all users' })
  @ApiQuery({ name: 'page', type: 'number', example: 1 })
  @ApiQuery({ name: 'limit', type: 'number', example: 10 })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Users retrieved successfully',
  })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({ status: HttpStatus.FORBIDDEN, description: 'Forbidden' })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async findAll(
    @Query('page') page: number = 1,
    @Query('limit') limit: number = 10,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.findAll(page, limit, currentUser);
  }

  @Get(':userId')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Retrieve a user by ID' })
  @ApiParam({ name: 'userId', description: 'Unique ID of the user' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'User retrieved successfully',
  })
  @ApiResponse({ status: HttpStatus.NOT_FOUND, description: 'User not found' })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({ status: HttpStatus.FORBIDDEN, description: 'Forbidden' })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async findOne(
    @Param('userId') userId: string,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.findOne(userId, currentUser);
  }

  @UseGuards(AdminGuard)
  @Post()
  @HttpCode(HttpStatus.CREATED)
  @ApiOperation({ summary: 'Create a new user' })
  @ApiBody({ type: CreateUserDto })
  @ApiResponse({
    status: HttpStatus.CREATED,
    description: 'User created successfully',
  })
  @ApiResponse({
    status: HttpStatus.BAD_REQUEST,
    description: 'Invalid input data',
  })
  @ApiResponse({
    status: HttpStatus.CONFLICT,
    description: 'Email or phone number already exists',
  })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({ status: HttpStatus.FORBIDDEN, description: 'Forbidden' })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async create(
    @Body() createUserDto: CreateUserDto,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.createUser(createUserDto, currentUser);
  }

  @UseGuards(AuthenticatedGuard)
  @Put(':userId')
  @HttpCode(HttpStatus.ACCEPTED)
  @ApiOperation({ summary: 'Update an existing user' })
  @ApiParam({ name: 'userId', description: 'Unique ID of the user' })
  @ApiBody({ type: UserUpdateDto })
  @ApiResponse({
    status: HttpStatus.ACCEPTED,
    description: 'User updated successfully',
  })
  @ApiResponse({
    status: HttpStatus.BAD_REQUEST,
    description: 'Invalid input data',
  })
  @ApiResponse({ status: HttpStatus.NOT_FOUND, description: 'User not found' })
  @ApiResponse({
    status: HttpStatus.CONFLICT,
    description: 'Email or phone number already exists',
  })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({ status: HttpStatus.FORBIDDEN, description: 'Forbidden' })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async update(
    @Param('userId') userId: string,
    @Body() updateUserDto: UserUpdateDto,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.update(userId, updateUserDto, currentUser);
  }

  @UseGuards(AuthenticatedGuard)
  @Delete(':userId')
  @HttpCode(HttpStatus.NON_AUTHORITATIVE_INFORMATION)
  @ApiOperation({ summary: 'Delete a user' })
  @ApiParam({ name: 'userId', description: 'Unique ID of the user' })
  @ApiResponse({
    status: HttpStatus.NON_AUTHORITATIVE_INFORMATION,
    description: 'User deleted successfully',
  })
  @ApiResponse({ status: HttpStatus.NOT_FOUND, description: 'User not found' })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({ status: HttpStatus.FORBIDDEN, description: 'Forbidden' })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async remove(
    @Param('userId') userId: string,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.remove(userId, currentUser);
  }

  @UseGuards(AdminGuard)
  @Get(':userId/roles')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Retrieve roles for a user' })
  @ApiParam({ name: 'userId', description: 'Unique ID of the user' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Roles retrieved successfully',
  })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({ status: HttpStatus.FORBIDDEN, description: 'Forbidden' })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async getUserRoles(
    @Param('userId') userId: string,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.getUserRoles(userId);
  }

  @UseGuards(AuthenticatedGuard)
  @Post(':userId/settings')
  @HttpCode(HttpStatus.CREATED)
  @ApiOperation({ summary: 'Create a user setting' })
  @ApiParam({ name: 'userId', description: 'Unique ID of the user' })
  @ApiBody({ type: CreateUserSettingDto })
  @ApiResponse({
    status: HttpStatus.CREATED,
    description: 'User setting created successfully',
  })
  @ApiResponse({
    status: HttpStatus.BAD_REQUEST,
    description: 'Invalid input data',
  })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({
    status: HttpStatus.FORBIDDEN,
    description: 'Forbidden (requires SuperAdmin role)',
  })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async createUserSetting(
    @Param('userId') userId: string,
    @Body() createUserSettingDto: CreateUserSettingDto,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.createUserSetting(
      userId,
      createUserSettingDto,
      currentUser,
    );
  }

  @UseGuards(AuthenticatedGuard)
  @Get(':userId/settings')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Retrieve all settings for a user' })
  @ApiParam({ name: 'userId', description: 'Unique ID of the user' })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'User settings retrieved successfully',
  })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({ status: HttpStatus.FORBIDDEN, description: 'Forbidden' })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async getUserSettings(
    @Param('userId') userId: string,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.getUserSettings(userId);
  }

  @UseGuards(AuthenticatedGuard)
  @Put(':userId/settings/:userSettingId')
  @HttpCode(HttpStatus.ACCEPTED)
  @ApiOperation({ summary: 'Update a user setting' })
  @ApiParam({ name: 'userId', description: 'Unique ID of the user' })
  @ApiParam({
    name: 'userSettingId',
    description: 'Unique ID of the user setting',
  })
  @ApiBody({ type: UpdateUserSettingDto })
  @ApiResponse({
    status: HttpStatus.ACCEPTED,
    description: 'User setting updated successfully',
  })
  @ApiResponse({
    status: HttpStatus.BAD_REQUEST,
    description: 'Invalid input data',
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'Setting not found',
  })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({ status: HttpStatus.FORBIDDEN, description: 'Forbidden' })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async updateUserSetting(
    @Param('userId') userId: string,
    @Param('userSettingId') userSettingId: string,
    @Body() updateUserSettingDto: UpdateUserSettingDto,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.updateUserSetting(
      userId,
      userSettingId,
      updateUserSettingDto,
      currentUser,
    );
  }

  @UseGuards(AuthenticatedGuard)
  @Delete(':userId/settings/:userSettingId')
  @HttpCode(HttpStatus.NON_AUTHORITATIVE_INFORMATION)
  @ApiOperation({ summary: 'Delete a user setting' })
  @ApiParam({ name: 'userId', description: 'Unique ID of the user' })
  @ApiParam({
    name: 'userSettingId',
    description: 'Unique ID of the user setting',
  })
  @ApiResponse({
    status: HttpStatus.NON_AUTHORITATIVE_INFORMATION,
    description: 'User setting deleted successfully',
  })
  @ApiResponse({
    status: HttpStatus.NON_AUTHORITATIVE_INFORMATION,
    description: 'Setting not found',
  })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({ status: HttpStatus.FORBIDDEN, description: 'Forbidden' })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async deleteUserSetting(
    @Param('userId') userId: string,
    @Param('userSettingId') userSettingId: string,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.deleteUserSetting(
      userId,
      userSettingId,
      currentUser,
    );
  }

  @UseGuards(AdminGuard)
  @Patch(':userId/status')
  @HttpCode(HttpStatus.ACCEPTED)
  @ApiOperation({ summary: 'Update user status' })
  @ApiParam({ name: 'userId', description: 'Unique ID of the user' })
  @ApiBody({ type: UpdateUserStatusDto })
  @ApiResponse({
    status: HttpStatus.ACCEPTED,
    description: 'User status updated successfully',
  })
  @ApiResponse({
    status: HttpStatus.BAD_REQUEST,
    description: 'Invalid input data',
  })
  @ApiResponse({ status: HttpStatus.NOT_FOUND, description: 'User not found' })
  @ApiResponse({ status: HttpStatus.UNAUTHORIZED, description: 'Unauthorized' })
  @ApiResponse({ status: HttpStatus.FORBIDDEN, description: 'Forbidden' })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error',
  })
  async updateUserStatus(
    @Param('userId') userId: string,
    @Body() updateStatusDto: UpdateUserStatusDto,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.updateUserStatus(
      userId,
      updateStatusDto.status,
      currentUser,
    );
  }

  @Post(':userId/subscribe')
  @UseGuards(AuthenticatedGuard)
  @HttpCode(HttpStatus.OK)
  @ApiOperation({
    summary: 'Subscribe user to a plan',
    description: 'Allows a user to subscribe to a monthly, yearly, or two-months-discount plan.',
  })
  @ApiParam({ name: 'userId', description: 'Unique ID of the user' })
  @ApiBody({ type: SubscribeToPlanDto })
  @ApiResponse({
    status: HttpStatus.OK,
    description: 'Subscription successful.',
  })
  @ApiResponse({
    status: HttpStatus.NOT_FOUND,
    description: 'User or wallet not found.',
  })
  @ApiResponse({
    status: HttpStatus.BAD_REQUEST,
    description: 'Invalid plan or insufficient wallet balance.',
  })
  @ApiResponse({
    status: HttpStatus.INTERNAL_SERVER_ERROR,
    description: 'Internal server error during subscription.',
  })
  async subscribeToPlan(
    @Param('userId') userId: string,
    @Body() subscribeDto: SubscribeToPlanDto,
    @CurrentUser() currentUser: ICurrentUser,
  ) {
    return this.usersService.subscribeToPlan(
      userId,
      subscribeDto.plan,
      subscribeDto.businessId,
      currentUser,
    );
  }
}
