  async findOne(
    userId: string,
    currentUser: ICurrentUser
  ) {
    try {
      return await this.db.transaction(async (tx) => {
        // Fetch user with related data
        const user = await tx.query.users.findFirst({
          where: eq(schema.users.userId, userId),
          columns: { password: false },
          with: {
            userAuth: true,
            wallet:true,
            business:true,
            userSettings: true,
            usersStats: true,
            userRoles: {
              with: {
                role: true,
              },
            },
          },
        });

        if (!user) {
          throw new NotFoundException(`User with ${userId} does not exist`);
        }

        // Check if current user follows the target user
        const following = await tx.query.follows.findFirst({
          where: sql`${schema.follows.followerId} = ${currentUser.userId} AND ${schema.follows.followingId} = ${userId}`,
        });

        return {
          ...user,
          following: !!following,
        };
      });
    } catch (error) {
      this.logger.error(`Error finding user: ${error.message}`, error.stack);
      this.databaseErrorService.handleDrizzleError(error, 'user retrieval');
    }
  }